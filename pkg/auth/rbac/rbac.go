package rbac

type Action string

type Scope string

type Permission struct {
	Action Action
	Scopes []Scope
}

func NewPermissionFromAction(action Action) Permission {
	return Permission{
		Action: action,
	}
}
