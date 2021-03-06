// +build linux freebsd

package daemon

import (
	"runtime"

	"github.com/sirupsen/logrus"
)

// getSize returns the real size & virtual size of the container.
func (daemon *Daemon) getSize(containerID string) (int64, int64) {
	var (
		sizeRw, sizeRootfs int64
		err                error
	)

	// Safe to index by runtime.GOOS as Unix hosts don't support multiple
	// container operating systems.
	rwlayer, err := daemon.layerStores[runtime.GOOS].GetRWLayer(containerID)
	if err != nil {
		logrus.Errorf("Failed to compute size of container rootfs %v: %v", containerID, err)
		return sizeRw, sizeRootfs
	}
	defer daemon.layerStores[runtime.GOOS].ReleaseRWLayer(rwlayer)

	sizeRw, err = rwlayer.Size()
	if err != nil {
		logrus.Errorf("Driver %s couldn't return diff size of container %s: %s",
			daemon.GraphDriverName(runtime.GOOS), containerID, err)
		// FIXME: GetSize should return an error. Not changing it now in case
		// there is a side-effect.
		sizeRw = -1
	}

	if parent := rwlayer.Parent(); parent != nil {
		sizeRootfs, err = parent.Size()
		if err != nil {
			sizeRootfs = -1
		} else if sizeRw != -1 {
			sizeRootfs += sizeRw
		}
	}
	return sizeRw, sizeRootfs
}
