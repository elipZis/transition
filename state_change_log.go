package transition

import (
	"strings"

	"github.com/qor/admin"
	"github.com/qor/audited"
	"github.com/qor/qor/resource"
	"github.com/qor/roles"
	"gorm.io/gorm"
)

// StateChangeLog a model that used to keep state change logs
type StateChangeLog struct {
	gorm.Model
	ReferTable string
	ReferID    string
	From       string
	To         string
	Note       string `sql:"size:1024"`
	audited.AuditedModel
}

// GenerateReferenceKey generate reference key used for change log
func GenerateReferenceKey(model interface{}, db *gorm.DB) string {
	var primaryValues []string

	stmt := &gorm.Statement{DB: db}
	_ = stmt.Parse(model)
	for _, field := range stmt.Schema.PrimaryFields {
		primaryValues = append(primaryValues, field.Name)
	}

	return strings.Join(primaryValues, "::")
}

// GetStateChangeLogs get state change logs
func GetStateChangeLogs(model interface{}, db *gorm.DB) []StateChangeLog {
	var changelogs []StateChangeLog

	stmt := &gorm.Statement{DB: db}
	_ = stmt.Parse(model)
	db.Where("refer_table = ? AND refer_id = ?", stmt.Schema.Table, GenerateReferenceKey(model, db)).Find(&changelogs)

	return changelogs
}

// GetLastStateChange gets last state change
func GetLastStateChange(model interface{}, db *gorm.DB) *StateChangeLog {
	var changelog StateChangeLog

	stmt := &gorm.Statement{DB: db}
	_ = stmt.Parse(model)
	db.Where("refer_table = ? AND refer_id = ?", stmt.Schema.Table, GenerateReferenceKey(model, db)).Last(&changelog)
	if changelog.To == "" {
		return nil
	}
	return &changelog
}

// ConfigureQorResource used to configure transition for qor admin
func (stageChangeLog *StateChangeLog) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		if res.Permission == nil {
			res.Permission = roles.Deny(roles.Update, roles.Anyone).Deny(roles.Create, roles.Anyone)
		} else {
			res.Permission = res.Permission.Deny(roles.Update, roles.Anyone).Deny(roles.Create, roles.Anyone)
		}
	}
}
