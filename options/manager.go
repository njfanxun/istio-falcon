package options

import (
	"time"
)

type ManagerOptions struct {
	LeaseDuration           time.Duration `json:"leaseDuration,omitempty" yaml:"leaseDuration,omitempty" mapstructure:"leaseDuration,omitempty"`
	RenewDeadline           time.Duration `json:"renewDeadline,omitempty" yaml:"renewDeadline,omitempty" mapstructure:"renewDeadline,omitempty"`
	RetryPeriod             time.Duration `json:"retryPeriod,omitempty" yaml:"retryPeriod,omitempty" mapstructure:"retryPeriod,omitempty"`
	LeaderElection          bool          `json:"leaderElection,omitempty" yaml:"leaderElection,omitempty" mapstructure:"leaderElection,omitempty"`
	LeaderElectionNamespace string        `json:"leaderElectionNamespace,omitempty" yaml:"leaderElectionNamespace,omitempty" mapstructure:"leaderElectionNamespace,omitempty"`
	LeaderElectionID        string        `json:"leaderElectionID,omitempty" yaml:"leaderElectionID,omitempty" mapstructure:"leaderElectionID,omitempty"`
}

func NewDefaultManagerOptions() *ManagerOptions {
	return &ManagerOptions{
		LeaseDuration:           30 * time.Second,
		RenewDeadline:           15 * time.Second,
		RetryPeriod:             5 * time.Second,
		LeaderElection:          true,
		LeaderElectionNamespace: "istio-system",
		LeaderElectionID:        "falcon-controller-manager-leader-election",
	}

}
func (o *ManagerOptions) Validate() []error {
	var errs []error

	return errs
}
