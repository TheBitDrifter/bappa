module github.com/TheBitDrifter/bappa/blueprint

go 1.24.1

// replace github.com/TheBitDrifter/bappa/tteokbokki => ../tteokbokki/

replace github.com/TheBitDrifter/bappa/environment => ../environment/

require (
	github.com/TheBitDrifter/bappa/environment v0.0.0-00010101000000-000000000000
	github.com/TheBitDrifter/bappa/tteokbokki v0.0.0-20250330212722-f989bbd448b4
	github.com/TheBitDrifter/bappa/warehouse v0.0.0-20250408214137-aae872bb6dfc
)

require (
	github.com/TheBitDrifter/bappa/table v0.0.0-20250408213844-3ea7eedb470f // indirect
	github.com/TheBitDrifter/bark v0.0.0-20250302175939-26104a815ed9 // indirect
	github.com/TheBitDrifter/mask v0.0.1-early-alpha.1 // indirect
	github.com/TheBitDrifter/util v0.0.0-20241102212109-342f4c0a810e // indirect
)
