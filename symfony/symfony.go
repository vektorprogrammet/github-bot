package symfony

import (
	"fmt"
	"strings"
	"github.com/vektorprogrammet/github-bot/cmd"
)

type Symfony struct {
	WorkingDirectory string
}

func (s Symfony) InstallDependencies() error {
	_, err := cmd.Execute("composer install -n", s.WorkingDirectory)
	return err
}

func (s Symfony) UpdateDependencies() (bool, error) {
	updateResult, err := cmd.Execute("composer update", s.WorkingDirectory)
	if err != nil {
		return false, err
	}
	return !strings.Contains(updateResult, "Nothing to install or update"), nil
}

func (s Symfony) CreateParametersFile(dbName string) error {
	if _, err := cmd.Execute("cp ../parameters.yml app/config/parameters.yml", s.WorkingDirectory); err != nil{
		return err
	}
	_, err := cmd.Execute(fmt.Sprintf("sed -i 's/dbname/%s/g' app/config/parameters.yml", dbName), s.WorkingDirectory)
	return err
}

func (s Symfony) CreateDatabase() error {
	if _, err := cmd.Execute("php app/console doctrine:database:create", s.WorkingDirectory); err != nil{
		return err
	}
	if _, err := cmd.Execute("php app/console doctrine:schema:create", s.WorkingDirectory); err != nil{
		return err
	}
	_, err := cmd.Execute("php app/console doctrine:migrations:version --add --all -n", s.WorkingDirectory)
	return err
}

func (s Symfony) DropDatabase() error {
	_, err := cmd.Execute("php app/console doctrine:database:drop --force", s.WorkingDirectory)
	return err
}

func (s Symfony) ExecuteMigrations() error {
	_, err := cmd.Execute("php app/console doctrine:migrations:migrate -n", s.WorkingDirectory)
	if err != nil {
		return err
	}
	return nil
}

func (s Symfony) CreateMigration() (bool, error) {
	migrationResult, err := cmd.Execute("php app/console doctrine:migrations:diff", s.WorkingDirectory)
	if err != nil {
		return false, err
	}

	return strings.Contains(migrationResult, "Generated new migration"), nil
}

func (s Symfony) FixCodeStyle() (bool, error) {
	codeStyleResult, err := cmd.Execute("npm run cs", s.WorkingDirectory)
	if err != nil {
		return false, err
	}

	return strings.Contains(codeStyleResult, "1)"), nil
}
