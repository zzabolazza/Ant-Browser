package browser

func RunningProfiles(profiles []Profile) []Profile {
	result := make([]Profile, 0)
	for _, profile := range profiles {
		if profile.Running {
			result = append(result, profile)
		}
	}
	return result
}
