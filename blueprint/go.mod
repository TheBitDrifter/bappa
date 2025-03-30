module github.com/TheBitDrifter/bappa/blueprint

go 1.24.1

require (
	github.com/TheBitDrifter/bappa/tteokbokki v0.0.0-00010101000000-000000000000
	github.com/TheBitDrifter/bappa/warehouse v0.0.0-00010101000000-000000000000
)

require (
	github.com/TheBitDrifter/bark v0.0.0-20250302175939-26104a815ed9 // indirect
	github.com/TheBitDrifter/mask v0.0.1-early-alpha.1 // indirect
	github.com/TheBitDrifter/table v0.0.0-20250315162738-9d26b0df5cd1 // indirect
	github.com/TheBitDrifter/util v0.0.0-20241102212109-342f4c0a810e // indirect
)

replace github.com/TheBitDrifter/bappa/tteokbokki => ../tteokbokki/

replace github.com/TheBitDrifter/bappa/warehouse => ../warehouse/

replace github.com/TheBitDrifter/bappa/table/ => ../table/
