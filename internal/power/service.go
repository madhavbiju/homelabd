package power

import (
	"context"
	"fmt"
	"os/exec"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// Reboot schedules a system reboot.
func (s *Service) Reboot(ctx context.Context) error {
	// Execute 'shutdown -r now' or similar based on standard Linux tools.
	// If the binary is not running as root, this requires sudoers configuration like:
	// homelabd ALL=(ALL) NOPASSWD: /sbin/shutdown
	cmd := exec.CommandContext(ctx, "sudo", "/sbin/shutdown", "-r", "now")
	
	// If "sudo" is not found, fallback to directly calling shutdown (useful if already root)
	if err := cmd.Run(); err != nil {
		fallback := exec.CommandContext(ctx, "/sbin/shutdown", "-r", "now")
		if err2 := fallback.Run(); err2 != nil {
			return fmt.Errorf("failed to reboot: sudo err=%v, fallback err=%v", err, err2)
		}
	}
	
	return nil
}

// Shutdown schedules a system power-off.
func (s *Service) Shutdown(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "sudo", "/sbin/shutdown", "-h", "now")
	
	if err := cmd.Run(); err != nil {
		fallback := exec.CommandContext(ctx, "/sbin/shutdown", "-h", "now")
		if err2 := fallback.Run(); err2 != nil {
			return fmt.Errorf("failed to shutdown: sudo err=%v, fallback err=%v", err, err2)
		}
	}

	return nil
}
