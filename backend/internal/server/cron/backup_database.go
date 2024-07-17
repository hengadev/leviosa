package cron

import (
	"fmt"
)

// Une fonction pour backer mes BDD. Faire les migrations en gros.
func backupDatabase() {
	// TODO: use goose to do some migration and send that somewhere, where I back the data
	fmt.Println("Ici on bien dormi en fait.")
}
