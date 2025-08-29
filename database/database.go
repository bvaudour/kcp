package database

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	"git.kaosx.ovh/benjamin/collection/concurrent"
)

// Database is the decoded structure
// of a json database of packages.
type Database struct {
	LastUpdate    time.Time `json:"last_update"`
	IgnoreRepos   []string  `json:"ignore_repos"`
	BrokenDepends []string  `json:"broken_depends"`
	Packages      `json:"packages"`
}

// New returns a new empty database initialized
// by repositories of the organzation to ignore.
func New(ignored ...string) Database {
	return Database{
		IgnoreRepos: ignored,
	}
}

// Decode decodes the given file to the database
func (db *Database) Decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(db)
}

// Encode encodes the database to json
// and write it to the given file.
func (db Database) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(db)
}

// Load decodes the file in the given path and
// returns the decoded database.
func Load(fpath string, ignored ...string) (db Database, err error) {
	var file *os.File
	if file, err = os.Open(fpath); err != nil {
		return
	}
	defer file.Close()

	db = New(ignored...)
	db.Decode(file)

	return
}

// Save writes the database into the file on the given path.
func Save(fpath string, db Database) error {
	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	return db.Encode(file)
}

// UpdateBroken updates the broken depends.
func (db *Database) UpdateBroken() {
	db.BrokenDepends = db.Packages.SearchBroken()
}

// UpdateRemote updates the database from the remote server.
// It returns a counter of the changes.
func (db *Database) UpdateRemote(connector Connector, debug bool) (counter Counter, err error) {
	// Étape 7: Mettre à jour db.LastUpdate avec la date/heure du début du traitement.
	startTime := time.Now()
	defer func() {
		if err == nil {
			db.LastUpdate = startTime
		}
	}()

	// Étape 1: Évaluer le nombre d'appels nécessaires.
	count, err := connector.CountPublcRepos()
	if err != nil {
		return counter, err
	}
	pages := count / defaultLimit
	if count%defaultLimit > 0 {
		pages++
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	stack := concurrent.NewSlice[Package]()
	errChan := make(chan error, pages)

	// Canal pour limiter le nombre de goroutines simultanées
	sem := make(chan struct{}, defaultRoutines)

	// Étape 2 & 3: Exécuter connector.GetPage en parallèle et traiter les réponses.
	for i := 1; i <= pages; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(page int) {
			defer wg.Done()
			defer func() { <-sem }()

			// Vérifier si une autre goroutine a déjà échoué
			select {
			case <-ctx.Done():
				return
			default:
			}

			pagePackages, pageErr := connector.GetPage(page, defaultLimit)
			if pageErr != nil {
				errChan <- pageErr
				cancel() // Annuler toutes les autres goroutines
				return
			}

			var pageWg sync.WaitGroup
			// Étape 4: Traiter chaque paquet.
			for _, p := range pagePackages {
				// Vérifier à nouveau avant de lancer une nouvelle sous-goroutine
				select {
				case <-ctx.Done():
					return
				default:
				}

				pageWg.Add(1)
				sem <- struct{}{}
				go func(pkg Package) {
					defer pageWg.Done()
					defer func() { <-sem }()

					// 4.1. Ignorer le paquet si nécessaire.
					if slices.Contains(db.IgnoreRepos, pkg.Name) {
						return
					}

					pkg.noChange = true
					// 4.2. Vérifier si le paquet a été mis à jour.
					if pkg.UpdatedAt.After(db.LastUpdate) {
						if file, err := pkg.GetPKGBUID(debug); err == nil {
							pkg.updateFromPKGBUILD(file)
							pkg.noChange = false
						} else if debug {
							log.Printf("Failed to get PKGBUILD for %s: %v", pkg.Name, err)
						}
					}

					// 4.3. Récupérer la version locale et ajouter à la liste.
					pkg.LocalVersion = pkg.GetLocaleVersion()
					stack.Append(pkg)
				}(p)
			}
			pageWg.Wait()
		}(i)
	}

	// Attendre la fin de toutes les goroutines ou une erreur
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		// Toutes les goroutines ont terminé sans erreur
	case err = <-errChan:
		// Une erreur est survenue, on l'a récupérée et on retourne
		return counter, err
	}

	// Étape 5: Parcourir la liste des paquets à traiter.
	remotePackages := Packages(stack.CloseData())
	newPackages := make(Packages, 0, len(remotePackages))
	for _, p := range remotePackages {
		localPkg, exists := db.Packages.Get(p.Name)

		if !exists {
			// 5.1. Le paquet n'existe pas dans la base de données.
			counter.Added++
			if p.noChange {
				if file, err := p.GetPKGBUID(debug); err == nil {
					p.updateFromPKGBUILD(file)
				} else if debug {
					log.Printf("Failed to get PKGBUILD for new package %s: %v", p.Name, err)
				}
			}
			newPackages.Push(p)
		} else {
			if p.noChange {
				// 5.2. Le paquet existe et noChange vaut true.
				p.updateFromPackage(localPkg)
				if p.LocalVersion != localPkg.LocalVersion {
					counter.Updated++
				}
			} else {
				// 5.3. Le paquet existe et noChange vaut false (mis à jour).
				counter.Updated++
			}
			newPackages.Push(p)
		}
	}

	// Étape 6: Parcourir les paquets locaux pour trouver les paquets supprimés.
	for _, localPkg := range db.Packages {
		if !remotePackages.Contains(localPkg.Name) {
			counter.Deleted++
		}
	}

	// Étape 7: Mettre à jour la base de données.
	db.Packages = newPackages

	return counter, nil
}

// Update checks if updates are available in the database.
func (db *Database) Update(connector Connector, debug bool) (counter Counter, err error) {
	if counter, err = db.UpdateRemote(connector, debug); err == nil {
		db.UpdateBroken()
	}

	return
}
