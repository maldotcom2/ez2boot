package ldap

import (
	"encoding/json"
)

func (r *Repository) getLdapConfig() (LdapConfig, error) {
	// TODO
	return LdapConfig{}, nil
}

func (r *Repository) getGroupMappings() ([]LdapGroupMapping, error) {
	rows, err := r.Base.DB.Query("SELECT ad_group, permissions FROM ldap_group_mappings")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var mappings []LdapGroupMapping
	for rows.Next() {
		var m LdapGroupMapping
		var raw string
		if err := rows.Scan(&m.ADGroup, &raw); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(raw), &m.Permissions); err != nil {
			return nil, err
		}

		mappings = append(mappings, m)
	}

	return mappings, nil
}
