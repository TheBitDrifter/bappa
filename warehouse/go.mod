module github.com/TheBitDrifter/bappa/warehouse

go 1.23.3

replace github.com/TheBitDrifter/bappa/blueprint => ../blueprint/

replace github.com/TheBitDrifter/bappa/warehouse => ../warehouse/

require (
	github.com/TheBitDrifter/bark v0.0.0-20250302175939-26104a815ed9
	github.com/TheBitDrifter/mask v0.0.1-early-alpha.1
	github.com/TheBitDrifter/table v0.0.0-20250315162738-9d26b0df5cd1
	github.com/TheBitDrifter/warehouse v0.0.1-early-alpha.1
)

require github.com/TheBitDrifter/util v0.0.0-20241102212109-342f4c0a810e // indirect
