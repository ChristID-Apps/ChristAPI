package requests

import "encoding/json"

type ListRolesRequest struct {
	ID     *int64 `query:"id"`
	SiteID *int64 `query:"siteId"`
}

type CreateRoleRequest struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	SiteID      *int64  `json:"site_id"`
}

type UpdateRoleRequest struct {
	Name        *string         `json:"name"`
	Code        *string         `json:"code"`
	Description *string         `json:"description"`
	Present     map[string]bool `json:"-"`
}

func (r *UpdateRoleRequest) UnmarshalJSON(data []byte) error {
	type alias UpdateRoleRequest
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decoded.Present = make(map[string]bool, len(fields))
	for field := range fields {
		decoded.Present[field] = true
	}
	*r = UpdateRoleRequest(decoded)
	return nil
}
