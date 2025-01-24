package ur

import (
	"fmt"
	"log/slog"
	"strings"
)

const (
	DEFAULT_ACCELERATION = 2.5
	DEFAULT_VELOCITY     = 12.0
)

const (
	MOVE_J = "movej"
	MOVE_L = "movel"
	MOVE_P = "movep"
)

type MoveCmd struct {
	PosA       URPosition
	ViaPos     []URPosition // Optional
	PosB       *URPosition  // Optional
	Iterations int
	Type       string

	Acceleration float64
	Velocity     float64
}

func (cmd *MoveCmd) String() string {
	if cmd.Type == "" {
		cmd.Type = MOVE_J
	}
	if cmd.Acceleration <= 0 {
		cmd.Acceleration = DEFAULT_ACCELERATION
	}
	if cmd.Velocity <= 0 {
		cmd.Velocity = DEFAULT_VELOCITY
	}

	if cmd.Iterations > 0 && cmd.PosB != nil {
		var loopSeq strings.Builder
		loopSeq.WriteString("i = 0\n")
		loopSeq.WriteString(fmt.Sprintf("  while i < %d:\n", cmd.Iterations))

		loopSeq.WriteString(fmt.Sprintf("  %s(%s, a=%f, v=%f)\n", cmd.Type, floatArrayToString(cmd.PosA[:]), cmd.Acceleration, cmd.Velocity))

		if cmd.ViaPos != nil {
			for _, viaPos := range cmd.ViaPos {
				loopSeq.WriteString(fmt.Sprintf("  %s(%s, a=%f, v=%f)\n", cmd.Type, floatArrayToString(viaPos[:]), cmd.Acceleration, cmd.Velocity))
			}
		}

		loopSeq.WriteString(fmt.Sprintf("  %s(%s, a=%f, v=%f)\n", cmd.Type, floatArrayToString(cmd.PosB[:]), cmd.Acceleration, cmd.Velocity))
		loopSeq.WriteString("  i = i + 1\n")

		loopSeq.WriteString("end\n")
		return loopSeq.String()
	}

	// Default single move
	var s = fmt.Sprintf("%s(%s, a=%f, v=%f)", cmd.Type, floatArrayToString(cmd.PosA[:]), cmd.Acceleration, cmd.Velocity)
	if cmd.ViaPos != nil {
		for _, viaPos := range cmd.ViaPos {
			s += fmt.Sprintf("\n%s(%s, a=%f, v=%f)", cmd.Type, floatArrayToString(viaPos[:]), cmd.Acceleration, cmd.Velocity)
		}
	}
	if cmd.PosB != nil {
		s += fmt.Sprintf("\n%s(%s, a=%f, v=%f)", cmd.Type, floatArrayToString(cmd.PosB[:]), cmd.Acceleration, cmd.Velocity)
	}
	return s
}

type MoveOptions struct {
	Acceleration float64
	Velocity     float64
}

// Default options
func defaultMoveOptions() MoveOptions {
	slog.Info("Using default move options", "acceleration", DEFAULT_ACCELERATION, "velocity", DEFAULT_VELOCITY)
	return MoveOptions{
		Acceleration: DEFAULT_ACCELERATION,
		Velocity:     DEFAULT_VELOCITY,
	}
}

// Option function type
type MoveOption func(*MoveOptions)

// Function to set acceleration
func WithAcceleration(acceleration float64) MoveOption {
	return func(opts *MoveOptions) {
		opts.Acceleration = acceleration
	}
}

// Function to set velocity
func WithVelocity(velocity float64) MoveOption {
	return func(opts *MoveOptions) {
		opts.Velocity = velocity
	}
}

func (c *URController) MoveJ(joints URPosition, options ...MoveOption) error {
	if len(joints) != 6 {
		return fmt.Errorf(ErrInvalidNumberOfJoints, len(joints))
	}

	opts := defaultMoveOptions()

	for _, opt := range options {
		opt(&opts)
	}

	cmd := MoveCmd{
		PosA:         joints,
		Acceleration: opts.Acceleration,
		Velocity:     opts.Velocity,
	}

	return c.SendCommand(cmd.String())
}

func (c *URController) MoveJSequence(cmds []MoveCmd) error {
	programSeq := "def move_sequence():\n"

	for _, cmd := range cmds {
		// pad the sequence with 2 spaces in front
		s := cmd.String()
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			lines[i] = "  " + line
		}
		programSeq += strings.Join(lines, "\n") + "\n"
	}

	programSeq += "end\n"

	// log.Println(programSeq)
	return c.SendCommand(programSeq)
}

func (c *URController) DoWork() error {
	positions := []MoveCmd{
		{
			PosA: c.Positions.Home,
			Type: MOVE_J,
		},
		{
			PosA:       c.Positions.OverDevice,
			PosB:       &c.Positions.PickDevice,
			Iterations: 3,
			Type:       MOVE_J,
		},
		{
			PosA: c.Positions.Home,
			Type: MOVE_J,
		},
		{
			PosA:       c.Positions.OverLetter,
			PosB:       &c.Positions.PickLetter,
			Iterations: 3,
			Type:       MOVE_J,
		},
		{
			PosA: c.Positions.Home,
			Type: MOVE_J,
		},
	}

	return c.MoveJSequence(positions)
}

// type JointData struct {
// 	PackageSize int
// 	PackageType uint8
// 	QActual     float64
// 	QTarget     float64
// 	QdActual    float64
// 	IActual     float32
// 	VActual     float32
// 	TMotor      float32
// 	TMicro      float32
// 	JointMode   uint8
// }

// // RobotStatePackageType contains constants for the packageType field
// const (
// 	RobotStatePackageTypeJointData = 1
// )

// func (c *URController) GetJointPositions() (JointData, error) {
// 	err := c.SendCommand("get_actual_joint_positions()")
// 	if err != nil {
// 		return JointData{}, err
// 	}

// 	b, err := c.ReadRawResp()
// 	if err != nil {
// 		return JointData{}, err
// 	}

// 	if len(b) < 46 { // Ensure the response is at least the expected size
// 		return JointData{}, fmt.Errorf("response too short, got %d bytes", len(b))
// 	}

// 	packageSize := binary.LittleEndian.Uint32(b[0:4])
// 	packageType := b[4]

// 	if packageType != RobotStatePackageTypeJointData {
// 		return JointData{}, fmt.Errorf("unexpected package type: %d", packageType)
// 	}

// 	qActual := math.Float64frombits(binary.LittleEndian.Uint64(b[5:13]))
// 	qTarget := math.Float64frombits(binary.LittleEndian.Uint64(b[13:21]))
// 	qdActual := math.Float64frombits(binary.LittleEndian.Uint64(b[21:29]))
// 	iActual := math.Float32frombits(binary.LittleEndian.Uint32(b[29:33]))
// 	vActual := math.Float32frombits(binary.LittleEndian.Uint32(b[33:37]))
// 	tMotor := math.Float32frombits(binary.LittleEndian.Uint32(b[37:41]))
// 	tMicro := math.Float32frombits(binary.LittleEndian.Uint32(b[41:45]))
// 	jointMode := b[45]

// 	// Validate parsed values
// 	if math.IsNaN(qActual) || math.IsNaN(qTarget) {
// 		return JointData{}, fmt.Errorf("received NaN value in joint positions")
// 	}

// 	return JointData{
// 		PackageSize: int(packageSize),
// 		PackageType: packageType,
// 		QActual:     qActual,
// 		QTarget:     qTarget,
// 		QdActual:    qdActual,
// 		IActual:     iActual,
// 		VActual:     vActual,
// 		TMotor:      tMotor,
// 		TMicro:      tMicro,
// 		JointMode:   jointMode,
// 	}, nil
// }
