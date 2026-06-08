package models

const (
	HakStoreID        = "h4kSt0r3aE"
	HakStoreRepo      = "https://github.com/aaronearles/nextui-hak-store"
	StorefrontJsonURL = "https://raw.githubusercontent.com/aaronearles/nextui-hak-store/refs/heads/main/storefront.json"
	GitHubRoot        = "https://github.com/"
	RawGHUC           = "https://raw.githubusercontent.com/"
	RefMainStub       = "/refs/heads/main/"
	PakJsonStub       = "pak.json"

	SDRoot              = "/mnt/SDCARD"
	UserdataDir         = ".userdata"
	HakStoreUserDataDir = "hak-store"
	ToolDir             = "Tools"
	EmulatorDir         = "Emus"
)

type Platform string

const TG5040 Platform = "tg5040"
const TG5050 Platform = "tg5050"
const MY355 Platform = "my355"
