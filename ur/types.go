package ur

const (
	RTDE_PROTOCOL_VERSION_1            = 1   // RTDE protocol version 1
	RTDE_PROTOCOL_VERSION_2            = 2   // RTDE protocol version 2
	RTDE_REQUEST_PROTOCOL_VERSION      = 86  // ascii V
	RTDE_GET_URCONTROL_VERSION         = 118 // ascii v
	RTDE_TEXT_MESSAGE                  = 77  // ascii M
	RTDE_DATA_PACKAGE                  = 85  // ascii U
	RTDE_CONTROL_PACKAGE_SETUP_OUTPUTS = 79  // ascii O
	RTDE_CONTROL_PACKAGE_SETUP_INPUTS  = 73  // ascii I
	RTDE_CONTROL_PACKAGE_START         = 83  // ascii S
	RTDE_CONTROL_PACKAGE_PAUSE         = 80  // ascii P
)

const (
	TYPE_VECTOR_6D = "VECTOR6D"
	TYPE_DOUBLE    = "DOUBLE"
	TYPE_UINT32    = "UINT32"
)

// defaults and max values
const (
	MAX_FREQ = 125.0
)

type Header struct {
	PkgSize uint16
	Cmd     uint8
}
