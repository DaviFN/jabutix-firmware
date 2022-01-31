package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	rpio "github.com/stianeikeland/go-rpio/v4"
)

const (
	// these numbers must represent pwm pins on the raspberry pi board
	LEFT_WHEEL_PIN       = 19
	RIGHT_WHEEL_PIN      = 12
	ACTUAL_PWM_FREQUENCY = 10000
	DUTY_CYCLE_TOTAL     = 100
)

type discreteMovementConfig struct {
	LeftWheelPower                       uint32
	RightWheelPower                      uint32
	TimeToWaitToTurnOffPower             uint32
	TimeToWaitAfterPowerHasBeenTurnedOff uint32
	saveFileName                         string
}

type DiscreteMovement struct {
	mvConfig discreteMovementConfig
}

func (mv *DiscreteMovement) save() {
	mvConfigJsonStr, err := json.Marshal(mv.mvConfig)
	if err != nil {
		panic(err)
	}
	f, err2 := os.Create(mv.mvConfig.saveFileName)
	if err2 != nil {
		panic(err2)
	}
	_, err3 := f.WriteString(string(mvConfigJsonStr))
	if err3 != nil {
		panic(err3)
	}
}

func (mv *DiscreteMovement) load() {
	file, err := ioutil.ReadFile(mv.mvConfig.saveFileName)
	if err != nil {
		return
	}
	err2 := json.Unmarshal([]byte(file), &mv.mvConfig)
	if err2 != nil {
		panic(err)
	}
}

func (mv *DiscreteMovement) GetAsString() string {
	meAsString := fmt.Sprintf("%f-%f-%d-%d", 100*float32(mv.mvConfig.LeftWheelPower)/DUTY_CYCLE_TOTAL, 100*float32(mv.mvConfig.RightWheelPower)/DUTY_CYCLE_TOTAL, mv.mvConfig.TimeToWaitToTurnOffPower, mv.mvConfig.TimeToWaitAfterPowerHasBeenTurnedOff)
	return meAsString
}

func (mv *DiscreteMovement) SetParams(p1 uint32, p2 uint32, p3 uint32, p4 uint32) {
	mv.mvConfig.SetParams(p1, p2, p3, p4)
}

func (mvConfig *discreteMovementConfig) Print() {
	fmt.Printf("LeftWheelPower: %d\n", mvConfig.LeftWheelPower)
	fmt.Printf("RightWheelPower: %d\n", mvConfig.RightWheelPower)
	fmt.Printf("TimeToWaitToTurnOffPower: %d\n", mvConfig.TimeToWaitToTurnOffPower)
	fmt.Printf("TimeToWaitAfterPowerHasBeenTurnedOff: %d\n", mvConfig.TimeToWaitAfterPowerHasBeenTurnedOff)
}

func (mvConfig *discreteMovementConfig) SetParams(p1 uint32, p2 uint32, p3 uint32, p4 uint32) {
	mvConfig.LeftWheelPower = p1
	mvConfig.RightWheelPower = p2
	mvConfig.TimeToWaitToTurnOffPower = p3
	mvConfig.TimeToWaitAfterPowerHasBeenTurnedOff = p4
}

var leftWheel rpio.Pin
var rightWheel rpio.Pin

var ForwardMovement DiscreteMovement
var LeftMovement DiscreteMovement
var RightMovement DiscreteMovement

func initializeMovementConfigs() {
	fmt.Printf("initializing ForwardMovement...\n")
	ForwardMovement.mvConfig.SetParams(50, 50, 1000, 1000)
	ForwardMovement.mvConfig.saveFileName = "forwardMovement.json"
	ForwardMovement.load()
	ForwardMovement.mvConfig.Print()

	fmt.Printf("initializing LeftMovement...\n")
	LeftMovement.mvConfig.SetParams(50, 0, 1000, 1000)
	LeftMovement.mvConfig.saveFileName = "leftMovement.json"
	LeftMovement.load()
	LeftMovement.mvConfig.Print()

	fmt.Printf("initializing RightMovement...\n")
	RightMovement.mvConfig.SetParams(0, 50, 1000, 1000)
	RightMovement.mvConfig.saveFileName = "rightMovement.json"
	RightMovement.load()
	RightMovement.mvConfig.Print()

	fmt.Printf("movement configurations loaded\n")
}

func initializeWheelPins() {

	// obs: output frequency is calculated as being the pwm clock frequency
	//divided by cycle length, so ACTUAL_PWM_FREQUENCY will be the actual output frequency

	// pwm is controlled using DutyCycle with first parameter ranging from 0 to DUTY_CYCLE_TOTAL

	leftWheel = rpio.Pin(LEFT_WHEEL_PIN)
	leftWheel.Mode(rpio.Pwm)
	leftWheel.Freq(DUTY_CYCLE_TOTAL * ACTUAL_PWM_FREQUENCY)
	leftWheel.DutyCycle(0, DUTY_CYCLE_TOTAL)

	rightWheel = rpio.Pin(RIGHT_WHEEL_PIN)
	rightWheel.Mode(rpio.Pwm)
	rightWheel.Freq(DUTY_CYCLE_TOTAL * ACTUAL_PWM_FREQUENCY)
	rightWheel.DutyCycle(0, DUTY_CYCLE_TOTAL)

	fmt.Printf("wheel pins initialized\n")
}

func InitRpio() {
	err := rpio.Open()
	if err != nil {
		panic(err)
	}

	fmt.Printf("go-rpio initialized\n")

	initializeWheelPins()
	initializeMovementConfigs()
}

func moveWheels(mv DiscreteMovement) {
	leftWheel.DutyCycle(mv.mvConfig.LeftWheelPower, DUTY_CYCLE_TOTAL)
	rightWheel.DutyCycle(mv.mvConfig.RightWheelPower, DUTY_CYCLE_TOTAL)
	time.Sleep(time.Duration(mv.mvConfig.TimeToWaitToTurnOffPower) * time.Millisecond)
	leftWheel.DutyCycle(0, DUTY_CYCLE_TOTAL)
	rightWheel.DutyCycle(0, DUTY_CYCLE_TOTAL)
	time.Sleep(time.Duration(mv.mvConfig.TimeToWaitAfterPowerHasBeenTurnedOff) * time.Millisecond)
}

func MoveForward() {
	fmt.Printf("moving forward...\n")
	moveWheels(ForwardMovement)
	fmt.Printf("moved forward\n")
}

func MoveLeft() {
	fmt.Printf("moving diagonally to the left...\n")
	moveWheels(LeftMovement)
	fmt.Printf("moved diagonally to the left\n")
}

func MoveRight() {
	fmt.Printf("moving diagonally to the right...\n")
	moveWheels(RightMovement)
	fmt.Printf("moved diagonally to the right\n")
}

func GetDiscreteMovementConfigMessageString() string {
	return "f-" + ForwardMovement.GetAsString() + "-" + LeftMovement.GetAsString() + "-" + RightMovement.GetAsString()
}

func SetDiscreteMovementConfigBasedOnMessage(message string) {
	tokens := strings.Split(message, "-")

	param1, _ := strconv.ParseFloat(tokens[1], 32)
	param1 /= 100
	param1 *= DUTY_CYCLE_TOTAL
	param2, _ := strconv.ParseFloat(tokens[2], 32)
	param2 /= 100
	param2 *= DUTY_CYCLE_TOTAL
	param3, _ := strconv.Atoi(tokens[3])
	param4, _ := strconv.Atoi(tokens[4])

	param5, _ := strconv.ParseFloat(tokens[5], 32)
	param5 /= 100
	param5 *= DUTY_CYCLE_TOTAL
	param6, _ := strconv.ParseFloat(tokens[6], 32)
	param6 /= 100
	param6 *= DUTY_CYCLE_TOTAL
	param7, _ := strconv.Atoi(tokens[7])
	param8, _ := strconv.Atoi(tokens[8])

	param9, _ := strconv.ParseFloat(tokens[9], 32)
	param9 /= 100
	param9 *= DUTY_CYCLE_TOTAL
	param10, _ := strconv.ParseFloat(tokens[10], 32)
	param10 /= 100
	param10 *= DUTY_CYCLE_TOTAL
	param11, _ := strconv.Atoi(tokens[11])
	param12, _ := strconv.Atoi(tokens[12])

	ForwardMovement.SetParams(uint32(param1), uint32(param2), uint32(param3), uint32(param4))
	LeftMovement.SetParams(uint32(param5), uint32(param6), uint32(param7), uint32(param8))
	RightMovement.SetParams(uint32(param9), uint32(param10), uint32(param11), uint32(param12))
}
