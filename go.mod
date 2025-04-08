module github.com/TheBitDrifter/bappa

go 1.24.1

replace github.com/TheBitDrifter/bappa/coldbrew => ./coldbrew/

replace github.com/TheBitDrifter/bappa/blueprint => ./blueprint/

replace github.com/TheBitDrifter/bappa/tteokbokki => ./tteokbokki/

replace github.com/TheBitDrifter/bappa/table => ./table/

require (
	github.com/TheBitDrifter/bappa/table v0.0.0-20250406132441-e0111efa839a // indirect
	github.com/TheBitDrifter/mask v0.0.0-20241104160006-d17f4de74b8e // indirect
	github.com/TheBitDrifter/util v0.0.0-20241102212109-342f4c0a810e // indirect
)
