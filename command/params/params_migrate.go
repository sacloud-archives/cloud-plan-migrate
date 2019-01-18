package params

// MigrateMigrateParam is input parameters for the sacloud API
type MigrateMigrateParam struct {
	Selector      []string `json:"selector"`
	Assumeyes     bool     `json:"assumeyes"`
	CleanupDisk   bool     `json:"cleanup-disk"`
	DisableReboot bool     `json:"disable-reboot"`
	ID            int64    `json:"id"`
	IDs           []int64
}

// NewMigrateMigrateParam return new MigrateMigrateParam
func NewMigrateMigrateParam() *MigrateMigrateParam {
	return &MigrateMigrateParam{}
}

// Validate checks current values in model
func (p *MigrateMigrateParam) Validate() []error {
	errors := []error{}
	{
		validator := validateSakuraID
		errs := validator("--id", p.ID)
		if errs != nil {
			errors = append(errors, errs...)
		}
	}
	return errors
}

func (p *MigrateMigrateParam) SetSelector(v []string) {
	p.Selector = v
}

func (p *MigrateMigrateParam) GetSelector() []string {
	return p.Selector
}
func (p *MigrateMigrateParam) SetAssumeyes(v bool) {
	p.Assumeyes = v
}

func (p *MigrateMigrateParam) GetAssumeyes() bool {
	return p.Assumeyes
}
func (p *MigrateMigrateParam) SetCleanupDisk(v bool) {
	p.CleanupDisk = v
}

func (p *MigrateMigrateParam) GetCleanupDisk() bool {
	return p.CleanupDisk
}
func (p *MigrateMigrateParam) SetDisableReboot(v bool) {
	p.DisableReboot = v
}

func (p *MigrateMigrateParam) GetDisableReboot() bool {
	return p.DisableReboot
}
func (p *MigrateMigrateParam) SetID(v int64) {
	p.ID = v
}

func (p *MigrateMigrateParam) GetID() int64 {
	return p.ID
}
