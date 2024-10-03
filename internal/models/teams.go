package models

// TODO: IST Timezones

type Member struct {
	Member string
	Share  int
}

func (m *ApplicationModel) GetTeamByRefNo(sporic_ref_no string) ([]Member, error) {

	var members []Member

	rows, err := m.Db.Query("select member_name, share from team where sporic_ref_no = ?", sporic_ref_no)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var member Member
		err := rows.Scan(&member.Member, &member.Share)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	return members, nil
}
