# DakotaPost Portal Automation Client — Spec

## Background

`portal.dakotapost.net` is a white-labeled Laravel/PHP SaaS by **Tritek** that DakotaPost uses
as their customer portal. It is fully server-side rendered (no XHR/API — HTML pages only),
with jQuery ajax calls for in-page actions. Auth is session cookie + Laravel CSRF token
(`_token`), which is embedded in every page's JS and is stable for the lifetime of a session.

The goal is a Python library + CLI that:
1. Authenticates with the portal
2. Fetches and parses mail inbox data into structured Python objects
3. Downloads envelope images and scanned PDFs
4. Triggers actions (open & scan, forward, shred) programmatically
5. Exposes a clean CLI for daily use

---

## Tech Stack

- **Language**: Python 3.11+
- **HTTP**: `httpx` (sync client, cookie jar handles session automatically)
- **HTML parsing**: `BeautifulSoup4` with `lxml` parser
- **CLI**: `typer`
- **Config/secrets**: `.env` file via `python-dotenv` (never hardcode credentials)
- **Output formatting**: `rich` (tables, panels)
- **PDF/image download**: stream via `httpx` to disk

No async needed — this is a personal automation tool with low concurrency requirements.

---

## Project Structure

```
dakotapost/
├── .env.example
├── pyproject.toml
├── dakotapost/
│   ├── __init__.py
│   ├── client.py        # DakotaPostClient — auth + raw HTTP
│   ├── parser.py        # HTML → dataclass parsing
│   ├── models.py        # dataclasses: MailItem, Address, etc.
│   └── cli.py           # typer CLI entrypoint
└── tests/
    └── test_parser.py   # unit tests against saved HTML fixtures
```

---

## Configuration

`.env` file (gitignored):
```
DAKOTAPOST_EMAIL=your@email.com
DAKOTAPOST_PASSWORD=yourpassword
DAKOTAPOST_BASE_URL=https://portal.dakotapost.net
```

Load with `python-dotenv` at startup. Raise a clear error if any var is missing.

---

## Data Models (`models.py`)

```python
@dataclass
class MailItem:
    mail_id: int                  # data-mail-id attribute on <tr>
    date: str                     # "June 11, 2026"
    tracking_number: str          # e.g. "20989896738"
    image_url: str                # /customer/getfileContent?url=/mail_uploaded/...
    pdf_url: str | None           # None if not yet scanned
    status: str                   # "unprocessed" | "open_scan_requested" | "opened_and_scanned" | "forwarding_requested" | "shredded"
    address_id: int | None        # if forwarding is set
    delivery_method: str | None

@dataclass
class Address:
    address_id: int
    label: str                    # full display string from <option>
    address_type: str             # "domestic" | "international"

@dataclass
class AccountInfo:
    name: str
    pmb: str
    wallet_balance: str
    plan_renews: str
    status: str                   # "Open" | "Closed" etc.
```

---

## Client (`client.py`)

```python
class DakotaPostClient:
    BASE_URL = "https://portal.dakotapost.net"

    def __init__(self, email: str, password: str): ...

    # --- Auth ---
    def login(self) -> None:
        """POST credentials, establish session, extract and store _token."""

    def _get_csrf_token(self, html: str) -> str:
        """Parse _token value from page JS. Pattern: _token:'<value>'"""

    def _refresh_token(self) -> None:
        """GET /customer/dashboard and re-extract _token if needed."""

    # --- Pages ---
    def get_inbox_html(self) -> str:
        """GET /customer/dashboard — raw HTML."""

    def get_undecided_html(self) -> str:
        """GET /customer/undecided"""

    def get_open_scan_queue_html(self) -> str:
        """GET /customer/mailtoOpenAndScan"""

    # --- File downloads ---
    def download_file(self, url_path: str, dest: Path) -> Path:
        """GET /customer/getfileContent?url={url_path}, stream to dest."""

    # --- Actions (all POST with _token) ---
    def request_open_scan(self, mail_ids: list[int]) -> dict:
        """POST /customer/mail/open_scan/action"""

    def request_forward(
        self,
        mail_ids: list[int],
        address_id: int,
        delivery_method: int,
        send_date: str,          # "MM/DD/YYYY"
        notes: str = "",
    ) -> dict:
        """POST /customer/mail/physical_piece/action"""

    def request_shred(self, mail_ids: list[int]) -> dict:
        """POST /customer/mail/shred/action"""

    def undo_action(self, mail_ids: list[int]) -> dict:
        """POST /customer/mail/undo/action"""

    def get_delivery_methods(self, address_type: str) -> str:
        """GET /customer/get/physical/deliveries/{address_type} — returns HTML <option> list."""

    def email_scan(self, mail_id: int, email_address: str) -> dict:
        """GET /customer/actionSendMailToCustomer/{mail_id}?email_address={email}"""
```

### Auth Flow Detail

1. GET `/customer/dashboard` — will redirect to login page if not authenticated.
   Detect redirect to login by checking response URL or page title.
2. Scrape the login form's CSRF token from the login page HTML
   (Laravel login forms have a hidden `<input name="_token">`).
3. POST to `/customer/login` (or wherever the form action points) with:
   - `_token`: from login page
   - `email`: from env
   - `password`: from env
4. Follow redirect. If landed on dashboard, auth succeeded.
5. Extract the JS-embedded `_token` from the dashboard HTML for use in action POSTs.
   Pattern to match: `_token:'XXXXXXXXXX'` or `"_token":"XXXXXXXXXX"` in `<script>` tags.
6. Store session cookies in `httpx.Client` cookie jar (automatic).

If any action returns a 302 to the login page, re-authenticate and retry once.

### Action POST format

All action POSTs send `application/x-www-form-urlencoded` with:
```
_token=<csrf_token>
mailsToAction={"mailIds":[1830748, 1825022]}
```
Plus action-specific fields (address, delivery_method, etc.).

Responses are JSON: `{"error": false}` on success or `{"error": true, "msg": "..."}` on failure.

---

## Parser (`parser.py`)

```python
def parse_inbox(html: str) -> list[MailItem]:
    """
    Parse dashboard HTML into MailItem list.

    Each mail item is a <tr data-mail-id="XXXXXX"> in the main table.
    Extract:
      - mail_id: tr['data-mail-id']
      - date: .table_time label text
      - tracking_number: text node after <br> following .table_time
      - image_url: img.mail_image_path[data-src]
      - pdf_url: a[download] href (present only if scanned)
      - status: infer from classes/text present in .table_action_sec
        - has "OPENED AND SCANNED (sent)" → "opened_and_scanned"
        - has .open_and_scan_single link → "unprocessed"
        - etc.
    """

def parse_addresses(html: str) -> list[Address]:
    """Parse #sel_address <option> elements."""

def parse_account_info(html: str) -> AccountInfo:
    """Parse sidebar: .customer_name h3, .customer_status, wallet span, PMB, renews."""
```

---

## CLI (`cli.py`)

```
dakotapost inbox                   # list all mail, rich table
dakotapost inbox --status unprocessed

dakotapost download <mail_id>      # download envelope image + PDF (if available)
dakotapost download --all          # download all scanned PDFs

dakotapost scan <mail_id> [mail_id ...]     # request open & scan
dakotapost forward <mail_id> [mail_id ...]  # interactive: pick address, method, date
dakotapost shred <mail_id> [mail_id ...]    # request shred (prompts for confirmation)
dakotapost email <mail_id>                  # email PDF scan to yourself

dakotapost account                 # show account info, wallet, PMB
dakotapost addresses               # list saved addresses
```

`inbox` table columns: `ID | Date | Tracking # | Status | Has PDF`

All destructive actions (shred, forward) prompt `[y/N]` before proceeding unless `--yes` flag is passed.

---

## Known Endpoints (reference)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/customer/dashboard` | Inbox |
| GET | `/customer/undecided` | Undecided mail |
| GET | `/customer/mailtoOpenAndScan` | Open & scan queue |
| GET | `/customer/mailtoaddress` | Forwarded mail |
| GET | `/customer/mailtoshred` | Shredded |
| GET | `/customer/pickedup/mails` | Picked up |
| GET | `/customer/address/list` | Address management page |
| GET | `/customer/tracking/information` | Tracking |
| GET | `/customer/notification` | Notifications |
| GET | `/customer/wallet/amount` | Wallet page |
| GET | `/customer/getfileContent?url=...` | Download image or PDF |
| GET | `/customer/get/physical/deliveries/{type}` | Delivery method options |
| GET | `/customer/open_scan_single/{id}/mail` | Single item open & scan |
| GET | `/customer/shred_single/{id}/mail` | Single item shred |
| GET | `/customer/actionSendMailToCustomer/{id}?email_address=...` | Email scan |
| GET | `/customer/logout` | Logout |
| POST | `/customer/mail/open_scan/action` | Batch open & scan |
| POST | `/customer/mail/shred/action` | Batch shred |
| POST | `/customer/mail/physical_piece/action` | Batch forward |
| POST | `/customer/mail/physical_piece/action/revert` | Modify forward |
| POST | `/customer/mail/undo/action` | Batch undo |
| POST | `/customer/cancel/previous/request` | Cancel pending request |
| POST | `/addnew/customer/address/ajax` | Add domestic address |
| POST | `/addnew/customer/international/address/ajax` | Add international address |
| POST | `/customer/copymail/action/{folderId}` | Copy mail to folder |
| POST | `/customers/notification/removed` | Dismiss notification |

---

## Error Handling

- On login failure: raise `AuthenticationError` with message.
- On action `{"error": true, "msg": "..."}`: raise `PortalError(msg)`.
- On session expiry (redirect to login mid-session): re-auth once, retry, then raise.
- On HTTP errors (5xx): raise with status code and URL.
- Never silently swallow errors.

---

## Notes & Caveats

- The `_token` in the page JS is the **action** CSRF token (for AJAX POSTs), not the login
  form token. They are different — scrape each from the appropriate page.
- The login form's hidden `_token` input must be read fresh each time (it rotates on each
  login page load).
- Mail image URLs use the pattern `/mail_uploaded/YYYYMMDD/00{tracking_number}.jpg`.
  PDF URLs follow `/pdf/{timestamp}-{email}_{date}_{time}.pdf`.
- Delivery method IDs (integers): 1=First Class, 2=Priority, 3=Express, 4-8=FedEx domestic,
  39+=international. Load dynamically via `/customer/get/physical/deliveries/{type}` to
  get the current list.
- The portal has no pagination visible in the source — it appears to return all mail on
  one page. Verify this holds for large inboxes.
- `httpx.Client` should be used as a context manager and reused across requests within
  a session (do not create a new client per request).
