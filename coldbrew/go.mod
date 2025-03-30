module github.com/TheBitDrifter/bappa/coldbrew

go 1.24.1

replace github.com/TheBitDrifter/bappa/blueprint => ../blueprint/

replace github.com/TheBitDrifter/bappa/tteokbokki => ../tteokbokki

replace github.com/TheBitDrifter/bappa/warehouse => ../warehouse/

replace github.com/TheBitDrifter/bappa/table => ../table

require (
	github.com/TheBitDrifter/bappa/blueprint v0.0.0-00010101000000-000000000000
	github.com/TheBitDrifter/bappa/table v0.0.0-20250315162738-9d26b0df5cd1
	github.com/TheBitDrifter/bappa/tteokbokki v0.0.0-00010101000000-000000000000
	github.com/TheBitDrifter/bappa/warehouse v0.0.0-00010101000000-000000000000
	github.com/TheBitDrifter/bark v0.0.0-20250302175939-26104a815ed9
	github.com/TheBitDrifter/mask v0.0.1-early-alpha.1
	github.com/hajimehoshi/ebiten/v2 v2.8.7
	golang.org/x/image v0.25.0
)

require (
	github.com/TheBitDrifter/util v0.0.0-20241102212109-342f4c0a810e // indirect
	github.com/ebitengine/gomobile v0.0.0-20240911145611-4856209ac325 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/oto/v3 v3.3.3 // indirect
	github.com/ebitengine/purego v0.8.0 // indirect
	github.com/go-text/typesetting v0.2.0 // indirect
	github.com/jezek/xgb v1.1.1 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)
