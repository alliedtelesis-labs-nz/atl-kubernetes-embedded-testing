package generate

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"testrunner/pkg/config"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// Job generates a job manifest
func Job(cfg config.Config, namespace string) (*batchv1.Job, error) {
	hostProjectRoot := filepath.Join(cfg.WorkspacePath, cfg.ProjectRoot)
	if cfg.ProjectRoot == "." {
		hostProjectRoot = cfg.WorkspacePath
	}

	workingDir, err := calculateWorkingDirectory(cfg.ProjectRoot, cfg.WorkspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate working directory: %w", err)
	}

	// Get current user's UID and GID
	uid, gid, err := getCurrentUserIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user IDs: %w", err)
	}

	projectName := "project"
	if cfg.ProjectRoot == "." {
		if cwd, err := os.Getwd(); err == nil {
			projectName = filepath.Base(cwd)
		}
	} else {
		projectName = filepath.Base(cfg.ProjectRoot)
	}

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("ket-%s", projectName),
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          &cfg.BackoffLimit,
			ActiveDeadlineSeconds: &cfg.ActiveDeadlineS,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "default",
					RestartPolicy:      corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						{
							Name: "source-code",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: hostProjectRoot,
									Type: &[]corev1.HostPathType{corev1.HostPathDirectory}[0],
								},
							},
						},
						{
							Name: "reports",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:          pointer.Int64(uid),
						RunAsGroup:         pointer.Int64(gid),
						FSGroup:            pointer.Int64(gid),
						SupplementalGroups: []int64{gid},
					},
					Containers: []corev1.Container{
						{
							Name:            "test-runner",
							Image:           cfg.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/bin/sh",
								"-c",
								cfg.TestCommand,
							},
							WorkingDir: workingDir,
							Env: []corev1.EnvVar{
								{
									Name:  "KET_TEST_NAMESPACE",
									Value: namespace,
								},
								{
									Name:  "KET_PROJECT_ROOT",
									Value: cfg.ProjectRoot,
								},
								{
									Name:  "KET_WORKSPACE_PATH",
									Value: cfg.WorkspacePath,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "source-code",
									MountPath: "/workspace",
								},
								{
									Name:      "reports",
									MountPath: "/reports",
								},
							},
						},
					},
				},
			},
		},
	}

	return job, nil
}

// calculateWorkingDirectory calculates the working directory for the test runner
func calculateWorkingDirectory(projectRoot, workspacePath string) (string, error) {
	if projectRoot == "." {
		return workspacePath, nil
	}
	return filepath.Join(workspacePath, projectRoot), nil
}

// getCurrentUserIDs returns the current user's UID and GID
func getCurrentUserIDs() (int64, int64, error) {
	currentUser, err := user.Current()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get current user: %w", err)
	}

	uid, err := strconv.ParseInt(currentUser.Uid, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse UID: %w", err)
	}

	gid, err := strconv.ParseInt(currentUser.Gid, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse GID: %w", err)
	}

	return uid, gid, nil
}
