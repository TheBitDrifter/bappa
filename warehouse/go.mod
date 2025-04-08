module github.com/TheBitDrifter/bappa/warehouse

go 1.24.1

// replace github.com/TheBitDrifter/bappa/blueprint => ../blueprint

replace github.com/TheBitDrifter/bappa/table => ../table

require (
	github.com/TheBitDrifter/bappa/table v0.0.0-20250406131439-a591f228f237
	github.com/TheBitDrifter/bark v0.0.0-20250302175939-26104a815ed9
	github.com/TheBitDrifter/mask v0.0.1-early-alpha.1
)

require github.com/TheBitDrifter/util v0.0.0-20241102212109-342f4c0a810e // indirect
