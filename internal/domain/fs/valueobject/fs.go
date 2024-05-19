package valueobject

import "os"

func isWrite(flag int) bool {
	return flag&os.O_RDWR != 0 || flag&os.O_WRONLY != 0
}
