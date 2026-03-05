package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		collections, err := app.FindAllCollections()
		if err != nil {
			return err
		}

		for _, collection := range collections {
			if collection.Type != "base" && collection.Type != "auth" {
				continue
			}

			var hasCreated, hasUpdated bool
			for _, f := range collection.Fields {
				if f.GetName() == "created" {
					hasCreated = true
				}
				if f.GetName() == "updated" {
					hasUpdated = true
				}
			}

			if hasCreated && hasUpdated {
				continue
			}

			// Add missing fields
			if !hasCreated {
				collection.Fields.Add(&core.AutodateField{
					Name:     "created",
					OnCreate: true,
				})
			}
			if !hasUpdated {
				collection.Fields.Add(&core.AutodateField{
					Name:     "updated",
					OnCreate: true,
					OnUpdate: true,
				})
			}

			if err := app.Save(collection); err != nil {
				// Don't fail hard, just log error? Or fail?
				// Better to fail so we know it didn't work.
				return fmt.Errorf("failed to save collection %s: %w", collection.Name, err)
			}
		}

		return nil
	}, func(app core.App) error {
		// Revert logic (optional, usually skipped for fixes like this)
		return nil
	})
}
