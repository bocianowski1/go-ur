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

type MoveOption func(*MoveOptions)

func WithAcceleration(acceleration float64) MoveOption {
	return func(opts *MoveOptions) {
		opts.Acceleration = acceleration
	}
}

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
		s := cmd.String()
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			lines[i] = "  " + line
		}
		programSeq += strings.Join(lines, "\n") + "\n"
	}

	programSeq += "end\n"

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
