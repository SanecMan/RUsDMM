package app

import (
	"log"
	"runtime"
	"sdmm/app/selfupdate"
	"sdmm/env"
	"sdmm/req"
)

var remoteManifest selfupdate.Manifest

func (a *app) checkForUpdates() {
	log.Println("[app] checking for self updates...")

	manifest, err := selfupdate.FetchRemoteManifest()
	if err != nil {
		log.Printf("[app] unable to fetch remote manifest: %v", err)
		return
	}

	remoteManifest = manifest

	if manifest.Version == env.Version {
		log.Println("[app] application is up to date!")
		return
	}
	log.Println("[app] new update available:", manifest.Version)

	a.menu.SetUpdateAvailable(manifest.Version, manifest.Description)

	if remoteManifest.ForceUpdate {
		a.selfUpdate()
	}
}

func (a *app) selfUpdate() {
	a.menu.SetUpdating()

	var updateDownloadLink string

	switch runtime.GOOS {
	case "windows":
		updateDownloadLink = remoteManifest.DownloadLinks.Windows
	case "linux":
		updateDownloadLink = remoteManifest.DownloadLinks.Linux
	case "darwin":
		updateDownloadLink = remoteManifest.DownloadLinks.MacOS
	}

	log.Println("[app] updating with:", updateDownloadLink)

	go func() {
		latestUpdate, err := req.Get(updateDownloadLink)
		if err != nil {
			log.Println("[app] unable to get latest update:", err)
			a.menu.SetUpdateError()
			return
		}

		if err = selfupdate.Update(latestUpdate); err != nil {
			log.Println("[app] unable to complete self update:", err)
			a.menu.SetUpdateError()
			return
		}

		a.menu.SetUpdated()

		log.Println("[app] self update completed successfully!")
	}()
}