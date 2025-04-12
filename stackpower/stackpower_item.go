package stackpower

type StackPower struct {
	Name             string
	Mode             string
	Topology         string
	TotalPower       float64
	ReservedPower    float64
	AllocatedPower   float64
	UnusedPower      float64
	NumSwitches      int
	NumPowerSupplies int
}