package opticalflow

import (
	"elevatorproject/src/config"
	"time"
)

type PID struct {
	kp, ki, kd    float64
	integralLimit float64
	integral      float64
	previousError float64
	hasPrevious   bool
}

func CreateDefaultPID() PID {
	return PID{
		kp:            config.OpticalFlowKp,
		ki:            config.OpticalFlowKi,
		kd:            config.OpticalFlowKd,
		integralLimit: config.OpticalFlowIntegralLimit,
	}
}

func (p *PID) Update(setpoint, measurement float64, dt time.Duration) float64 {
	if dt <= 0 {
		return 0
	}
	errorValue := setpoint - measurement
	dtSeconds := dt.Seconds()
	p.integral += errorValue * dtSeconds
	if p.integral > p.integralLimit {
		p.integral = p.integralLimit
	}
	if p.integral < -p.integralLimit {
		p.integral = -p.integralLimit
	}

	derivative := 0.0
	if p.hasPrevious {
		derivative = (errorValue - p.previousError) / dtSeconds
	}

	p.previousError = errorValue
	p.hasPrevious = true

	return p.kp*errorValue + p.ki*p.integral + p.kd*derivative
}

func Run(sensorInput <-chan float64, pidOutput chan<- float64) {
	pid := CreateDefaultPID()
	lastUpdate := time.Now()

	for measurement := range sensorInput {
		now := time.Now()
		correction := pid.Update(config.OpticalFlowSetpoint, measurement, now.Sub(lastUpdate))
		lastUpdate = now

		if pidOutput == nil {
			continue
		}
		select {
		case pidOutput <- correction:
		default:
		}
	}
}
