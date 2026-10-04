package attendance

import "time"

func businessDate(at time.Time) (string, error) {
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return "", err
	}
	return at.In(jakarta).Format("2006-01-02"), nil
}
