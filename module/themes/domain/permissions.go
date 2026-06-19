package domain

// Permission identifies one grantable theme administration action.
type Permission string

const (
	// PermissionThemesView allows reading themes.
	PermissionThemesView Permission = "themes.view"
	// PermissionThemesImport allows importing packages.
	PermissionThemesImport Permission = "themes.import"
	// PermissionThemesEdit allows editing draft files and metadata.
	PermissionThemesEdit Permission = "themes.edit"
	// PermissionThemesValidate allows running validation.
	PermissionThemesValidate Permission = "themes.validate"
	// PermissionThemesPublish allows public or preview activation.
	PermissionThemesPublish Permission = "themes.publish"
	// PermissionThemesRollback allows rolling back theme activation.
	PermissionThemesRollback Permission = "themes.rollback"
	// PermissionThemesDelete allows deleting or archiving themes.
	PermissionThemesDelete Permission = "themes.delete"
	// PermissionThemesPreview allows previewing themes.
	PermissionThemesPreview Permission = "themes.preview"
	// PermissionThemesActivate allows changing active theme pointers.
	PermissionThemesActivate Permission = "themes.activate"
)

// ThemePermissions returns the built-in theme permission catalog.
func ThemePermissions() []Permission {
	return []Permission{
		PermissionThemesView,
		PermissionThemesImport,
		PermissionThemesEdit,
		PermissionThemesValidate,
		PermissionThemesPublish,
		PermissionThemesRollback,
		PermissionThemesDelete,
		PermissionThemesPreview,
		PermissionThemesActivate,
	}
}
