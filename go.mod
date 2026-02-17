module elevatorproject

require driver-go v0.0.0

replace driver-go => ./libs/driver-go

require Network-go v0.0.0

require github.com/google/uuid v1.6.0 // indirect

replace Network-go => ./libs/Network-go

go 1.25.5
