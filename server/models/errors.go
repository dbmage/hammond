package models

type VehicleAlreadyExistsError struct {
	Registration string
}

func (e *VehicleAlreadyExistsError) Error() string {
	return "vehicle with this url already exists"
}
