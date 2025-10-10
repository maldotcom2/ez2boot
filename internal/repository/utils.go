// Take in Unix epoch for converting to time, and new duration for parsing and adding
package repository

import "time"

func GetExpiryFromDuration(currentExpiry int64, duration string) (int64, error) {

	dur, err := time.ParseDuration(duration) // Parse string (eg, 4h) to time type
	if err != nil {
		return 0, err
	}

	// Case for new session
	if currentExpiry == 0 {
		now := time.Now().UTC()
		newExpiry := now.Add(dur).Unix()

		return newExpiry, nil

		// Case for update session
	} else {
		exp := time.Unix(currentExpiry, 0).UTC()
		newExpiry := exp.Add(dur).Unix()

		return newExpiry, nil
	}
}
