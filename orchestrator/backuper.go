package orchestrator

import (
	"log"
	"os"
	"runtime/pprof"
)

func NewBackuper(artifactManager ArtifactManager, logger Logger, deploymentManager DeploymentManager) *Backuper {
	return &Backuper{
		ArtifactManager:   artifactManager,
		Logger:            logger,
		DeploymentManager: deploymentManager,
	}
}

//go:generate counterfeiter -o fakes/fake_logger.go . Logger
type Logger interface {
	Debug(tag, msg string, args ...interface{})
	Info(tag, msg string, args ...interface{})
	Warn(tag, msg string, args ...interface{})
	Error(tag, msg string, args ...interface{})
}

//go:generate counterfeiter -o fakes/fake_deployment_manager.go . DeploymentManager
type DeploymentManager interface {
	Find(deploymentName string) (Deployment, error)
	SaveManifest(deploymentName string, artifact Artifact) error
}

type Backuper struct {
	ArtifactManager
	Logger

	DeploymentManager
}

type AuthInfo struct {
	Type   string
	UaaUrl string
}

//Backup checks if a deployment has backupable instances and backs them up.
func (b Backuper) Backup(deploymentName string) Error {
	f, err := os.Create("profile.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	bw := newBackupWorkflow(b, deploymentName)

	return bw.Run()
}

func (b Backuper) CanBeBackedUp(deploymentName string) (bool, Error) {
	bw := newBackupCheckWorkflow(b, deploymentName)

	err := bw.Run()
	return err == nil, err
}

func cleanupAndReturnErrors(d Deployment, err error) error {
	cleanupErr := d.Cleanup()
	if cleanupErr != nil {
		return Error{cleanupErr, err}
	}
	return err
}
