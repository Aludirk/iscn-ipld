package data

import (
	"fmt"
)

// ValidateParent between version and parent CID
func ValidateParent(version *Number, parent *Cid) error {
	ver, err := version.GetUint64()
	if err != nil {
		return err
	}

	if ver == 1 {
		if parent.IsDefined() {
			return fmt.Errorf("Parent should not be set as version <= 1")
		}
	} else if ver > 1 {
		if !parent.IsDefined() {
			return fmt.Errorf("Parent missed as version > 1")
		}
	}

	return nil
}
