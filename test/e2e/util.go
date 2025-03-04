/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Copyright 2019, 2021 The Multi-Cluster App Dispatcher Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package e2e

import (
	gcontext "context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	arbv1 "github.com/project-codeflare/multi-cluster-app-dispatcher/pkg/apis/controller/v1beta1"
	versioned "github.com/project-codeflare/multi-cluster-app-dispatcher/pkg/client/clientset/controller-versioned"
	csapi "github.com/project-codeflare/multi-cluster-app-dispatcher/pkg/controller/clusterstate/api"
)

var ninetySeconds = 90 * time.Second
var threeMinutes = 180 * time.Second
var tenMinutes = 600 * time.Second
var threeHundredSeconds = 300 * time.Second

var oneCPU = v1.ResourceList{"cpu": resource.MustParse("1000m")}
var twoCPU = v1.ResourceList{"cpu": resource.MustParse("2000m")}
var threeCPU = v1.ResourceList{"cpu": resource.MustParse("3000m")}

const (
	workerPriority = "worker-pri"
	masterPriority = "master-pri"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

type context struct {
	kubeclient *kubernetes.Clientset
	karclient  *versioned.Clientset

	namespace              string
	queues                 []string
	enableNamespaceAsQueue bool
}

func initTestContext() *context {
	enableNamespaceAsQueue, _ := strconv.ParseBool(os.Getenv("ENABLE_NAMESPACES_AS_QUEUE"))
	cxt := &context{
		namespace: "test",
		queues:    []string{"q1", "q2"},
	}

	home := homeDir()
	Expect(home).NotTo(Equal(""))

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	Expect(err).NotTo(HaveOccurred())

	cxt.karclient = versioned.NewForConfigOrDie(config)
	cxt.kubeclient = kubernetes.NewForConfigOrDie(config)

	cxt.enableNamespaceAsQueue = enableNamespaceAsQueue

	_, err = cxt.kubeclient.CoreV1().Namespaces().Create(gcontext.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cxt.namespace,
		},
	}, metav1.CreateOptions{})
	//Expect(err).NotTo(HaveOccurred())

	/* 	_, err = cxt.kubeclient.SchedulingV1beta1().PriorityClasses().Create(gcontext.Background(), &schedv1.PriorityClass{
	   		ObjectMeta: metav1.ObjectMeta{
	   			Name: masterPriority,
	   		},
	   		Value:         100,
	   		GlobalDefault: false,
	   	}, metav1.CreateOptions{})
	   	Expect(err).NotTo(HaveOccurred())

	   	_, err = cxt.kubeclient.SchedulingV1beta1().PriorityClasses().Create(gcontext.Background(), &schedv1.PriorityClass{
	   		ObjectMeta: metav1.ObjectMeta{
	   			Name: workerPriority,
	   		},
	   		Value:         1,
	   		GlobalDefault: false,
	   	}, metav1.CreateOptions{})
	   	Expect(err).NotTo(HaveOccurred()) */

	return cxt
}

func namespaceNotExist(ctx *context) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := ctx.kubeclient.CoreV1().Namespaces().Get(gcontext.Background(), ctx.namespace, metav1.GetOptions{})
		if !(err != nil && errors.IsNotFound(err)) {
			return false, err
		}
		return true, nil
	}
}

func cleanupTestContextExtendedTime(cxt *context, seconds time.Duration) {
	//foreground := metav1.DeletePropagationForeground
	/* err := cxt.kubeclient.CoreV1().Namespaces().Delete(gcontext.Background(), cxt.namespace, metav1.DeleteOptions{
		PropagationPolicy: &foreground,
	})
	Expect(err).NotTo(HaveOccurred()) */

	// err := cxt.kubeclient.SchedulingV1beta1().PriorityClasses().Delete(gcontext.Background(), masterPriority, metav1.DeleteOptions{
	// 	PropagationPolicy: &foreground,
	// })
	// Expect(err).NotTo(HaveOccurred())

	// err = cxt.kubeclient.SchedulingV1beta1().PriorityClasses().Delete(gcontext.Background(), workerPriority, metav1.DeleteOptions{
	// 	PropagationPolicy: &foreground,
	// })
	// Expect(err).NotTo(HaveOccurred())

	// Wait for namespace deleted.
	// err = wait.Poll(100*time.Millisecond, seconds, namespaceNotExist(cxt))
	// if err != nil {
	// 	fmt.Fprintf(os.Stdout, "[cleanupTestContextExtendedTime] Failure check for namespace: %s.\n", cxt.namespace)
	// }
	//Expect(err).NotTo(HaveOccurred())
}

func cleanupTestContext(cxt *context) {
	cleanupTestContextExtendedTime(cxt, ninetySeconds)
}

type taskSpec struct {
	min, rep int32
	img      string
	hostport int32
	req      v1.ResourceList
	affinity *v1.Affinity
	labels   map[string]string
}

type jobSpec struct {
	name      string
	namespace string
	queue     string
	tasks     []taskSpec
}

func getNS(context *context, job *jobSpec) string {
	if len(job.namespace) != 0 {
		return job.namespace
	}

	if context.enableNamespaceAsQueue {
		if len(job.queue) != 0 {
			return job.queue
		}
	}

	return context.namespace
}

func createGenericAWTimeoutWithStatus(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{
		"apiVersion": "batch/v1",
		"kind": "Job",
		"metadata": {
			"name": "aw-test-jobtimeout-with-comp-1",
			"namespace": "test"
		},
		"spec": {
			"completions": 1,
			"parallelism": 1,
			"template": {
				"metadata": {
					"labels": {
						"appwrapper.mcad.ibm.com": "aw-test-jobtimeout-with-comp-1"
					}
				},
				"spec": {
					"containers": [
						{
							"args": [
								"sleep infinity"
							],
							"command": [
								"/bin/bash",
								"-c",
								"--"
							],
							"image": "ubuntu:latest",
							"imagePullPolicy": "IfNotPresent",
							"name": "aw-test-jobtimeout-with-comp-1",
							"resources": {
								"limits": {
									"cpu": "100m",
									"memory": "256M"
								},
								"requests": {
									"cpu": "100m",
									"memory": "256M"
								}
							}
						}
					],
					"restartPolicy": "Never"
				}
			}
		}
	}`)
	var schedSpecMin int = 1
	var dispatchDurationSeconds int = 10
	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
				DispatchDuration: arbv1.DispatchDurationSpec{
					Limit: dispatchDurationSeconds,
				},
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-test-jobtimeout-with-comp-1-job"),
							Namespace: "test",
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
						CompletionStatus: "Complete",
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createJobEx(context *context, job *jobSpec) ([]*batchv1.Job, *arbv1.AppWrapper) {
	var jobs []*batchv1.Job
	var appwrapper *arbv1.AppWrapper
	var min int32

	ns := getNS(context, job)

	for i, task := range job.tasks {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", job.name, i),
				Namespace: ns,
			},
			Spec: batchv1.JobSpec{
				Parallelism: &task.rep,
				Completions: &task.rep,
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      task.labels,
						Annotations: map[string]string{arbv1.AppWrapperAnnotationKey: job.name},
					},
					Spec: v1.PodSpec{
						SchedulerName: "default",
						RestartPolicy: v1.RestartPolicyNever,
						Containers:    createContainers(task.img, task.req, task.hostport),
						Affinity:      task.affinity,
					},
				},
			},
		}

		job, err := context.kubeclient.BatchV1().Jobs(job.Namespace).Create(gcontext.Background(), job, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
		jobs = append(jobs, job)

		min = min + task.min
	}

	rb := []byte(`{"kind": "Pod", "apiVersion": "v1", "metadata": { "name": "foo"}}`)

	var schedSpecMin int = 1

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.name,
			Namespace: ns,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", job.name, "resource1"),
							Namespace: ns,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypePod,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(ns).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return jobs, appwrapper
}

/*
func taskPhase(ctx *context, pg *arbv1.PodGroup, phase []v1.PodPhase, taskNum int) wait.ConditionFunc {
	return func() (bool, error) {
		pg, err := ctx.karclient.Scheduling().PodGroups(pg.Namespace).Get(pg.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		pods, err := ctx.kubeclient.CoreV1().Pods(pg.Namespace).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		readyTaskNum := 0
		for _, pod := range pods.Items {
			if gn, found := pod.Annotations[arbv1.GroupNameAnnotationKey]; !found || gn != pg.Name {
				continue
			}

			for _, p := range phase {
				if pod.Status.Phase == p {
					readyTaskNum++
					break
				}
			}
		}

		return taskNum <= readyTaskNum, nil
	}
}
*/

func anyPodsExist(ctx *context, awNamespace string, awName string) wait.ConditionFunc {
	return func() (bool, error) {
		podList, err := ctx.kubeclient.CoreV1().Pods(awNamespace).List(gcontext.Background(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		podExistsNum := 0
		for _, podFromPodList := range podList.Items {

			// First find a pod from the list that is part of the AW
			if awn, found := podFromPodList.Labels["appwrapper.mcad.ibm.com"]; !found || awn != awName {
				//DEBUG fmt.Fprintf(os.Stdout, "[anyPodsExist] Pod %s in phase: %s not part of AppWrapper: %s, labels: %#v\n",
				//DEBUG 	podFromPodList.Name, podFromPodList.Status.Phase, awName, podFromPodList.Labels)
				continue
			}
			podExistsNum++
			fmt.Fprintf(os.Stdout, "[anyPodsExist] Found Pod %s in phase: %s as part of AppWrapper: %s, labels: %#v\n",
				podFromPodList.Name, podFromPodList.Status.Phase, awName, podFromPodList.Labels)
		}

		return podExistsNum > 0, nil
	}
}

func podPhase(ctx *context, awNamespace string, awName string, pods []*v1.Pod, phase []v1.PodPhase, taskNum int) wait.ConditionFunc {
	return func() (bool, error) {
		podList, err := ctx.kubeclient.CoreV1().Pods(awNamespace).List(gcontext.Background(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		if podList == nil || podList.Size() < 1 {
			fmt.Fprintf(os.Stdout, "[podPhase] Listing podList found for Namespace: %s/%s resulting in no podList found that could match AppWrapper with pod count: %d\n",
				awNamespace, awName, len(pods))
		}

		phaseListTaskNum := 0

		for _, podFromPodList := range podList.Items {

			// First find a pod from the list that is part of the AW
			if awn, found := podFromPodList.Labels["appwrapper.mcad.ibm.com"]; !found || awn != awName {
				fmt.Fprintf(os.Stdout, "[podPhase] Pod %s in phase: %s not part of AppWrapper: %s, labels: %#v\n",
					podFromPodList.Name, podFromPodList.Status.Phase, awName, podFromPodList.Labels)
				continue
			}

			// Next check to see if it is a phase we are looking for
			for _, p := range phase {

				// If we found the phase make sure it is part of the list of pod provided in the input
				if podFromPodList.Status.Phase == p {
					matchToPodsFromInput := false
					var inputPodIDs []string
					for _, inputPod := range pods {
						inputPodIDs = append(inputPodIDs, fmt.Sprintf("%s.%s", inputPod.Namespace, inputPod.Name))
						if strings.Compare(podFromPodList.Namespace, inputPod.Namespace) == 0 &&
							strings.Compare(podFromPodList.Name, inputPod.Name) == 0 {
							phaseListTaskNum++
							matchToPodsFromInput = true
							break
						}

					}
					if matchToPodsFromInput == false {
						fmt.Fprintf(os.Stdout, "[podPhase] Pod %s in phase: %s does not match any input pods: %#v \n",
							podFromPodList.Name, podFromPodList.Status.Phase, inputPodIDs)
					}
					break
				}
			}
		}

		return taskNum == phaseListTaskNum, nil
	}
}

func awStatePhase(ctx *context, aw *arbv1.AppWrapper, phase []arbv1.AppWrapperState, taskNum int, quite bool) wait.ConditionFunc {
	return func() (bool, error) {
		aw, err := ctx.karclient.ArbV1().AppWrappers(aw.Namespace).Get(aw.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		phaseCount := 0
		if !quite {
			fmt.Fprintf(os.Stdout, "[awStatePhase] AW %s found with state: %s.\n", aw.Name, aw.Status.State)
		}

		for _, p := range phase {
			if aw.Status.State == p {
				phaseCount++
				break
			}
		}
		return 1 <= phaseCount, nil
	}
}

func cleanupTestObjectsPtr(context *context, appwrappersPtr *[]*arbv1.AppWrapper) {
	cleanupTestObjectsPtrVerbose(context, appwrappersPtr, true)
}

func cleanupTestObjectsPtrVerbose(context *context, appwrappersPtr *[]*arbv1.AppWrapper, verbose bool) {
	if appwrappersPtr == nil {
		fmt.Fprintf(os.Stdout, "[cleanupTestObjectsPtr] No  AppWrappers to cleanup.\n")
	} else {
		cleanupTestObjects(context, *appwrappersPtr)
	}
}

func cleanupTestObjects(context *context, appwrappers []*arbv1.AppWrapper) {
	cleanupTestObjectsVerbose(context, appwrappers, true)
}

func cleanupTestObjectsVerbose(context *context, appwrappers []*arbv1.AppWrapper, verbose bool) {
	if appwrappers == nil {
		fmt.Fprintf(os.Stdout, "[cleanupTestObjects] No AppWrappers to cleanup.\n")
		return
	}

	for _, aw := range appwrappers {
		//context.karclient.ArbV1().AppWrappers(context.namespace).Delete(aw.Name, &metav1.DeleteOptions{PropagationPolicy: &foreground})

		pods := getPodsOfAppWrapper(context, aw)
		awNamespace := aw.Namespace
		awName := aw.Name
		fmt.Fprintf(os.Stdout, "[cleanupTestObjects] Deleting AW %s.\n", aw.Name)
		err := deleteAppWrapper(context, aw.Name)
		Expect(err).NotTo(HaveOccurred())

		// Wait for the pods of the deleted the appwrapper to be destroyed
		for _, pod := range pods {
			fmt.Fprintf(os.Stdout, "[cleanupTestObjects] Awaiting pod %s/%s to be deleted for AW %s.\n",
				pod.Namespace, pod.Name, aw.Name)
		}
		err = waitAWPodsDeleted(context, awNamespace, awName, pods)

		// Final check to see if pod exists
		if err != nil {
			var podsStillExisting []*v1.Pod
			for _, pod := range pods {
				podExist, _ := context.kubeclient.CoreV1().Pods(pod.Namespace).Get(gcontext.Background(), pod.Name, metav1.GetOptions{})
				if podExist != nil {
					fmt.Fprintf(os.Stdout, "[cleanupTestObjects] Found pod %s/%s %s, not completedly deleted for AW %s.\n", podExist.Namespace, podExist.Name, podExist.Status.Phase, aw.Name)
					podsStillExisting = append(podsStillExisting, podExist)
				}
			}
			if len(podsStillExisting) > 0 {
				err = waitAWPodsDeleted(context, awNamespace, awName, podsStillExisting)
			}
		}
		Expect(err).NotTo(HaveOccurred())
	}
	cleanupTestContext(context)
}

func awPodPhase(ctx *context, aw *arbv1.AppWrapper, phase []v1.PodPhase, taskNum int, quite bool) wait.ConditionFunc {
	return func() (bool, error) {
		defer GinkgoRecover()

		aw, err := ctx.karclient.ArbV1().AppWrappers(aw.Namespace).Get(aw.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		podList, err := ctx.kubeclient.CoreV1().Pods(aw.Namespace).List(gcontext.Background(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		if podList == nil || podList.Size() < 1 {
			fmt.Fprintf(os.Stdout, "[awPodPhase] Listing podList found for Namespace: %s resulting in no podList found that could match AppWrapper: %s \n",
				aw.Namespace, aw.Name)
		}

		readyTaskNum := 0
		for _, pod := range podList.Items {
			if awn, found := pod.Labels["appwrapper.mcad.ibm.com"]; !found || awn != aw.Name {
				if !quite {
					fmt.Fprintf(os.Stdout, "[awPodPhase] Pod %s not part of AppWrapper: %s, labels: %s\n", pod.Name, aw.Name, pod.Labels)
				}
				continue
			}

			for _, p := range phase {
				if pod.Status.Phase == p {
					//DEBUGif quite {
					//DEBUG	fmt.Fprintf(os.Stdout, "[awPodPhase] Found pod %s of AppWrapper: %s, phase: %v\n", pod.Name, aw.Name, p)
					//DEBUG}
					readyTaskNum++
					break
				} else {
					pMsg := pod.Status.Message
					if len(pMsg) > 0 {
						pReason := pod.Status.Reason
						fmt.Fprintf(os.Stdout, "[awPodPhase] pod: %s, phase: %s, reason: %s, message: %s\n", pod.Name, p, pReason, pMsg)
					}
					containerStatuses := pod.Status.ContainerStatuses
					for _, containerStatus := range containerStatuses {
						waitingState := containerStatus.State.Waiting
						if waitingState != nil {
							wMsg := waitingState.Message
							if len(wMsg) > 0 {
								wReason := waitingState.Reason
								containerName := containerStatus.Name
								fmt.Fprintf(os.Stdout, "[awPodPhase] condition for pod: %s, phase: %s, container name: %s, "+
									"reason: %s, message: %s\n", pod.Name, p, containerName, wReason, wMsg)
							}
						}
					}
				}
			}
		}

		//DEBUGif taskNum <= readyTaskNum && quite {
		//DEBUG	fmt.Fprintf(os.Stdout, "[awPodPhase] Successfully found %v podList of AppWrapper: %s, state: %s\n", readyTaskNum, aw.Name, aw.Status.State)
		//DEBUG}

		return taskNum <= readyTaskNum, nil
	}
}

/*
func podGroupUnschedulable(ctx *context, pg *arbv1.PodGroup, time time.Time) wait.ConditionFunc {
	return func() (bool, error) {
		pg, err := ctx.karclient.Scheduling().PodGroups(pg.Namespace).Get(pg.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		events, err := ctx.kubeclient.CoreV1().Events(pg.Namespace).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		for _, event := range events.Items {
			target := event.InvolvedObject
			if target.Name == pg.Name && target.Namespace == pg.Namespace {
				if event.Reason == string(arbv1.UnschedulableEvent) && event.LastTimestamp.After(time) {
					return true, nil
				}
			}
		}

		return false, nil
	}
}
*/
/*
func waitPodGroupReady(ctx *context, pg *arbv1.PodGroup) error {
	return waitTasksReadyEx(ctx, pg, int(pg.Spec.MinMember))
}

func waitPodGroupPending(ctx *context, pg *arbv1.PodGroup) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, taskPhase(ctx, pg,
		[]v1.PodPhase{v1.PodPending}, int(pg.Spec.MinMember)))
}

func waitTasksReadyEx(ctx *context, pg *arbv1.PodGroup, taskNum int) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, taskPhase(ctx, pg,
		[]v1.PodPhase{v1.PodRunning, v1.PodSucceeded}, taskNum))
}

func waitTasksPendingEx(ctx *context, pg *arbv1.PodGroup, taskNum int) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, taskPhase(ctx, pg,
		[]v1.PodPhase{v1.PodPending}, taskNum))
}

func waitPodGroupUnschedulable(ctx *context, pg *arbv1.PodGroup) error {
	now := time.Now()
	return wait.Poll(10*time.Second, ninetySeconds, podGroupUnschedulable(ctx, pg, now))
}
*/

func waitAWNonComputeResourceActive(ctx *context, aw *arbv1.AppWrapper) error {
	return waitAWNamespaceActive(ctx, aw)
}

func waitAWNamespaceActive(ctx *context, aw *arbv1.AppWrapper) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, awNamespacePhase(ctx, aw,
		[]v1.NamespacePhase{v1.NamespaceActive}))
}

func awNamespacePhase(ctx *context, aw *arbv1.AppWrapper, phase []v1.NamespacePhase) wait.ConditionFunc {
	return func() (bool, error) {
		aw, err := ctx.karclient.ArbV1().AppWrappers(aw.Namespace).Get(aw.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		namespaces, err := ctx.kubeclient.CoreV1().Namespaces().List(gcontext.Background(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		readyTaskNum := 0
		for _, namespace := range namespaces.Items {
			if awns, found := namespace.Labels["appwrapper.mcad.ibm.com"]; !found || awns != aw.Name {
				continue
			}

			for _, p := range phase {
				if namespace.Status.Phase == p {
					readyTaskNum++
					break
				}
			}
		}

		return 0 < readyTaskNum, nil
	}
}

func waitAWPodsReady(ctx *context, aw *arbv1.AppWrapper) error {
	return waitAWPodsReadyEx(ctx, aw, int(aw.Spec.SchedSpec.MinAvailable), false)
}

func waitAWPodsCompleted(ctx *context, aw *arbv1.AppWrapper, timeout time.Duration) error {
	return waitAWPodsCompletedEx(ctx, aw, int(aw.Spec.SchedSpec.MinAvailable), false, timeout)
}

func waitAWPodsNotCompleted(ctx *context, aw *arbv1.AppWrapper) error {
	return waitAWPodsNotCompletedEx(ctx, aw, int(aw.Spec.SchedSpec.MinAvailable), false)
}

func waitAWReadyQuiet(ctx *context, aw *arbv1.AppWrapper) error {
	return waitAWPodsReadyEx(ctx, aw, int(aw.Spec.SchedSpec.MinAvailable), true)
}

func waitAWAnyPodsExists(ctx *context, aw *arbv1.AppWrapper) error {
	return waitAWPodsExists(ctx, aw, ninetySeconds)
}

func waitAWPodsExists(ctx *context, aw *arbv1.AppWrapper, timeout time.Duration) error {
	return wait.Poll(100*time.Millisecond, timeout, anyPodsExist(ctx, aw.Namespace, aw.Name))
}

func waitAWDeleted(ctx *context, aw *arbv1.AppWrapper, pods []*v1.Pod) error {
	return waitAWPodsTerminatedEx(ctx, aw.Namespace, aw.Name, pods, 0)
}

func waitAWPodsDeleted(ctx *context, awNamespace string, awName string, pods []*v1.Pod) error {
	return waitAWPodsDeletedVerbose(ctx, awNamespace, awName, pods, true)
}

func waitAWPodsDeletedVerbose(ctx *context, awNamespace string, awName string, pods []*v1.Pod, verbose bool) error {
	return waitAWPodsTerminatedEx(ctx, awNamespace, awName, pods, 0)
}

func waitAWPending(ctx *context, aw *arbv1.AppWrapper) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, awPodPhase(ctx, aw,
		[]v1.PodPhase{v1.PodPending}, int(aw.Spec.SchedSpec.MinAvailable), false))
}

func waitAWPodsReadyEx(ctx *context, aw *arbv1.AppWrapper, taskNum int, quite bool) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, awPodPhase(ctx, aw,
		[]v1.PodPhase{v1.PodRunning, v1.PodSucceeded}, taskNum, quite))
}

func waitAWPodsCompletedEx(ctx *context, aw *arbv1.AppWrapper, taskNum int, quite bool, timeout time.Duration ) error {
	return wait.Poll(100*time.Millisecond, timeout, awPodPhase(ctx, aw,
		[]v1.PodPhase{v1.PodSucceeded}, taskNum, quite))
}

func waitAWPodsNotCompletedEx(ctx *context, aw *arbv1.AppWrapper, taskNum int, quite bool) error {
	return wait.Poll(100*time.Millisecond, threeMinutes, awPodPhase(ctx, aw,
		[]v1.PodPhase{v1.PodPending, v1.PodRunning, v1.PodFailed, v1.PodUnknown}, taskNum, quite))
}

func waitAWPodsPending(ctx *context, aw *arbv1.AppWrapper) error {
	return waitAWPodsPendingEx(ctx, aw, int(aw.Spec.SchedSpec.MinAvailable), false)
}

func waitAWPodsPendingEx(ctx *context, aw *arbv1.AppWrapper, taskNum int, quite bool) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, awPodPhase(ctx, aw,
		[]v1.PodPhase{v1.PodPending}, taskNum, quite))
}

func waitAWPodsTerminatedEx(ctx *context, namespace string, name string, pods []*v1.Pod, taskNum int) error {
	return waitAWPodsTerminatedExVerbose(ctx, namespace, name, pods, taskNum, true)
}

func waitAWPodsTerminatedExVerbose(ctx *context, namespace string, name string, pods []*v1.Pod, taskNum int, verbose bool) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, podPhase(ctx, namespace, name, pods,
		[]v1.PodPhase{v1.PodRunning, v1.PodSucceeded, v1.PodUnknown, v1.PodFailed, v1.PodPending}, taskNum))
}

func createContainers(img string, req v1.ResourceList, hostport int32) []v1.Container {
	container := v1.Container{
		Image:           img,
		Name:            img,
		ImagePullPolicy: v1.PullIfNotPresent,
		Resources: v1.ResourceRequirements{
			Requests: req,
		},
	}

	if hostport > 0 {
		container.Ports = []v1.ContainerPort{
			{
				ContainerPort: hostport,
				HostPort:      hostport,
			},
		}
	}

	return []v1.Container{container}
}

func createReplicaSet(context *context, name string, rep int32, img string, req v1.ResourceList) *appv1.ReplicaSet {
	deploymentName := "deployment.k8s.io"
	deployment := &appv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: appv1.ReplicaSetSpec{
			Replicas: &rep,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					deploymentName: name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{deploymentName: name},
				},
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyAlways,
					Containers: []v1.Container{
						{
							Image:           img,
							Name:            name,
							ImagePullPolicy: v1.PullIfNotPresent,
							Resources: v1.ResourceRequirements{
								Requests: req,
							},
						},
					},
				},
			},
		},
	}

	deployment, err := context.kubeclient.AppsV1().ReplicaSets(context.namespace).Create(gcontext.Background(), deployment, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())

	return deployment
}

func createJobAWWithInitContainer(context *context, name string, requeuingTimeInSeconds int, requeuingGrowthType string, requeuingMaxNumRequeuings int ) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "batch/v1",
		"kind": "Job",
	"metadata": {
		"name": "aw-job-3-init-container",
		"namespace": "test",
		"labels": {
			"app": "aw-job-3-init-container"
		}
	},
	"spec": {
		"parallelism": 3,
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-job-3-init-container"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-job-3-init-container"
				}
			},
			"spec": {
				"terminationGracePeriodSeconds": 1,
				"restartPolicy": "Never",
				"initContainers": [
					{
						"name": "job-init-container",
						"image": "k8s.gcr.io/busybox:latest",
						"command": ["sleep", "200"],
						"resources": {
							"requests": {
								"cpu": "500m"
							}
						}
					}
				],
				"containers": [
					{
						"name": "job-container",
						"image": "k8s.gcr.io/busybox:latest",
						"command": ["sleep", "10"],
						"resources": {
							"requests": {
								"cpu": "500m"
							}
						}
					}
				]
			}
		}
	}} `)

	var minAvailable int = 3

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: minAvailable,
				Requeuing: arbv1.RequeuingTemplate{
					TimeInSeconds: requeuingTimeInSeconds,
					GrowthType: requeuingGrowthType,
					MaxNumRequeuings: requeuingMaxNumRequeuings,
				},
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: context.namespace,
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
						CompletionStatus: "Complete",
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-3",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-3"
		}
	},
	"spec": {
		"replicas": 3,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-3"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-3"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-3"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-3",
						"image": "k8s.gcr.io/echoserver:1.4",
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 3

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAWwith900CPU(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-900cpu",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-900cpu"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-900cpu"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-900cpu"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-900cpu"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-900cpu",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "900m"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAWwith550CPU(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-550cpu",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-550cpu"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-550cpu"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-550cpu"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-550cpu"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-550cpu",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "550m"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAWwith125CPU(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-125cpu",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-125cpu"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-125cpu"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-125cpu"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-125cpu"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-125cpu",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "125m"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAWwith126CPU(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-126cpu",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-126cpu"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-126cpu"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-126cpu"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-126cpu"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-126cpu",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "126m"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAWwith350CPU(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-350cpu",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-350cpu"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-350cpu"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-350cpu"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-350cpu"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-350cpu",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "350m"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAWwith351CPU(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-351cpu",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-351cpu"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-351cpu"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-351cpu"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-351cpu"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-351cpu",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "351m"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAWwith426CPU(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
	"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-426cpu",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-426cpu"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-426cpu"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-426cpu"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-426cpu"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-426cpu",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "426m"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createDeploymentAWwith425CPU(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
	"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-425cpu",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-425cpu"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-425cpu"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-425cpu"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-425cpu"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-425cpu",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "425m"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeDeployment,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericDeploymentAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-generic-deployment-3",
		"namespace": "test",
		"labels": {
			"app": "aw-generic-deployment-3"
		}
	},
	"spec": {
		"replicas": 3,
		"selector": {
			"matchLabels": {
				"app": "aw-generic-deployment-3"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-generic-deployment-3"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-generic-deployment-3"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-generic-deployment-3",
						"image": "k8s.gcr.io/echoserver:1.4",
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 3

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-generic-deployment-3-item1"),
							Namespace: context.namespace,
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
						CompletionStatus: "Progressing",
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericJobAWWithStatus(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{
		"apiVersion": "batch/v1",
		"kind": "Job",
		"metadata": {
			"name": "aw-test-job-with-comp-1",
			"namespace": "test"
		},
		"spec": {
			"completions": 1,
			"parallelism": 1,
			"template": {
				"metadata": {
					"labels": {
						"appwrapper.mcad.ibm.com": "aw-test-job-with-comp-1"
					}
				},
				"spec": {
					"containers": [
						{
							"args": [
								"sleep 5"
							],
							"command": [
								"/bin/bash",
								"-c",
								"--"
							],
							"image": "ubuntu:latest",
							"imagePullPolicy": "IfNotPresent",
							"name": "aw-test-job-with-comp-1",
							"resources": {
								"limits": {
									"cpu": "100m",
									"memory": "256M"
								},
								"requests": {
									"cpu": "100m",
									"memory": "256M"
								}
							}
						}
					],
					"restartPolicy": "Never"
				}
			}
		}
	}`)
	//var schedSpecMin int = 1

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				//MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-test-job-with-comp-1"),
							Namespace: "test",
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
						CompletionStatus: "Complete",
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericJobAWWithScheduleSpec(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{
		"apiVersion": "batch/v1",
		"kind": "Job",
		"metadata": {
			"name": "aw-test-job-with-scheduling-spec",
			"namespace": "test"
		},
		"spec": {
			"completions": 2,
			"parallelism": 2,
			"template": {
				"metadata": {
					"labels": {
						"appwrapper.mcad.ibm.com": "aw-test-job-with-scheduling-spec"
					}
				},
				"spec": {
					"containers": [
						{
							"command": [
								"/bin/bash",
								"-c",
								"--"
							],
							"args": [
								"sleep 5"
							],
							"image": "ubuntu:latest",
							"imagePullPolicy": "IfNotPresent",
							"name": "aw-test-job-with-scheduling-spec",
							"resources": {
								"limits": {
									"cpu": "100m",
									"memory": "256M"
								},
								"requests": {
									"cpu": "100m",
									"memory": "256M"
								}
							}
						}
					],
					"restartPolicy": "Never"
				}
			}
		}
	}`)

	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-test-job-with-scheduling-spec"),
							Namespace: "test",
						},
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericJobAWtWithLargeCompute(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{
		"apiVersion": "batch/v1",
		"kind": "Job",
		"metadata": {
			"name": "aw-test-job-with-large-comp-1",
			"namespace": "test"
		},
		"spec": {
			"completions": 1,
			"parallelism": 1,
			"template": {
				"metadata": {
					"labels": {
						"appwrapper.mcad.ibm.com": "aw-test-job-with-large-comp-1"
					}
				},
				"spec": {
					"containers": [
						{
							"args": [
								"sleep 5"
							],
							"command": [
								"/bin/bash",
								"-c",
								"--"
							],
							"image": "ubuntu:latest",
							"imagePullPolicy": "IfNotPresent",
							"name": "aw-test-job-with-comp-1",
							"resources": {
								"limits": {
									"cpu": "10000m",
									"memory": "256M",
									"nvidia.com/gpu": "100"
								},
								"requests": {
									"cpu": "100000m",
									"memory": "256M",
									"nvidia.com/gpu": "100"
								}
							}
						}
					],
					"restartPolicy": "Never"
				}
			}
		}
	}`)
	//var schedSpecMin int = 1

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				//MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-test-job-with-large-comp-1"),
							Namespace: "test",
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
						//CompletionStatus: "Complete",
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericServiceAWWithNoStatus(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{
		"apiVersion": "v1",
		"kind": "Service",
		"metadata": {
			"labels": {
				"appwrapper.mcad.ibm.com": "test-dep-job-item",
				"resourceName": "test-dep-job-item-svc"
			},
			"name": "test-dep-job-item-svc",
			"namespace": "test"
		},
		"spec": {
			"ports": [
				{
					"name": "client",
					"port": 10001,
					"protocol": "TCP",
					"targetPort": 10001
				},
				{
					"name": "dashboard",
					"port": 8265,
					"protocol": "TCP",
					"targetPort": 8265
				},
				{
					"name": "redis",
					"port": 6379,
					"protocol": "TCP",
					"targetPort": 6379
				}
			],
			"selector": {
				"component": "test-dep-job-item-svc"
			},
			"sessionAffinity": "None",
			"type": "ClusterIP"
		}
	}`)
	//var schedSpecMin int = 1

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				//MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-test-job-with-comp-1"),
							Namespace: "test",
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
						CompletionStatus: "Complete",
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericDeploymentAWWithMultipleItems(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-2-status",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-2-status"
		}
	},
	"spec": {
		"replicas": 1,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-2-status"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-2-status"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-2-status"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-2-status",
						"image": "k8s.gcr.io/echoserver:1.4",
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)

	rb1 := []byte(`{"apiVersion": "apps/v1",
	"kind": "Deployment", 
"metadata": {
	"name": "aw-deployment-3-status",
	"namespace": "test",
	"labels": {
		"app": "aw-deployment-3-status"
	}
},
"spec": {
	"replicas": 1,
	"selector": {
		"matchLabels": {
			"app": "aw-deployment-3-status"
		}
	},
	"template": {
		"metadata": {
			"labels": {
				"app": "aw-deployment-3-status"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-3-status"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "aw-deployment-3-status",
					"image": "k8s.gcr.io/echoserver:1.4",
					"ports": [
						{
							"containerPort": 80
						}
					]
				}
			]
		}
	}
}} `)

	var schedSpecMin int = 1

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-deployment-2-status"),
							Namespace: "test",
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
						CompletionStatus: "Progressing",
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-deployment-3-status"),
							Namespace: "test",
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb1,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericDeploymentAWWithService(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "Deployment", 
	"metadata": {
		"name": "aw-deployment-3-status",
		"namespace": "test",
		"labels": {
			"app": "aw-deployment-3-status"
		}
	},
	"spec": {
		"replicas": 1,
		"selector": {
			"matchLabels": {
				"app": "aw-deployment-3-status"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-deployment-3-status"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-deployment-3-status"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-deployment-3-status",
						"image": "k8s.gcr.io/echoserver:1.4",
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)

	rb1 := []byte(`{
		"apiVersion": "v1",
		"kind": "Service",
		"metadata": {
			"name": "my-service",
			"namespace": "test"
		},
		"spec": {
			"clusterIP": "10.96.76.247",
			"clusterIPs": [
				"10.96.76.247"
			],
			"ipFamilies": [
				"IPv4"
			],
			"ipFamilyPolicy": "SingleStack",
			"ports": [
				{
					"port": 80,
					"protocol": "TCP",
					"targetPort": 9376
				}
			],
			"selector": {
				"app.kubernetes.io/name": "MyApp"
			},
			"sessionAffinity": "None",
			"type": "ClusterIP"
		},
		"status": {
			"loadBalancer": {}
		}
	}`)

	var schedSpecMin int = 1

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "aw-deployment-3-status"),
							Namespace: "test",
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
						CompletionStatus: "Progressing",
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "my-service"),
							Namespace: "test",
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb1,
						},
						CompletionStatus: "bogus",
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericDeploymentWithCPUAW(context *context, name string, cpuDemand string, replicas int) *arbv1.AppWrapper {
	rb := []byte(fmt.Sprintf(`{
	"apiVersion": "apps/v1",
	"kind": "Deployment", 
	"metadata": {
		"name": "%s",
		"namespace": "test",
		"labels": {
			"app": "%s"
		}
	},
	"spec": {
		"replicas": %d,
		"selector": {
			"matchLabels": {
				"app": "%s"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "%s"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "%s"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "%s",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "%s"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `, name, name, replicas, name, name, name, name, cpuDemand))

	var schedSpecMin int = replicas

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						DesiredAvailable: 1,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericDeploymentCustomPodResourcesWithCPUAW(context *context, name string, customPodCpuDemand string, cpuDemand string, replicas int, requeuingTimeInSeconds int) *arbv1.AppWrapper {
	rb := []byte(fmt.Sprintf(`{
	"apiVersion": "apps/v1",
	"kind": "Deployment", 
	"metadata": {
		"name": "%s",
		"namespace": "test",
		"labels": {
			"app": "%s"
		}
	},
	"spec": {
		"replicas": %d,
		"selector": {
			"matchLabels": {
				"app": "%s"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "%s"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "%s"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "%s",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"requests": {
								"cpu": "%s"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `, name, name, replicas, name, name, name, name, cpuDemand))

	var schedSpecMin int = replicas
	var customCpuResource = v1.ResourceList{"cpu": resource.MustParse(customPodCpuDemand)}

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
				Requeuing: arbv1.RequeuingTemplate{
					TimeInSeconds: requeuingTimeInSeconds,
				},
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						CustomPodResources: []arbv1.CustomPodResourceTemplate{
							{
								Replicas: replicas,
								Requests: customCpuResource,
							},
						},
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createNamespaceAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "v1",
		"kind": "Namespace", 
	"metadata": {
		"name": "aw-namespace-0",
		"labels": {
			"app": "aw-namespace-0"
		}
	}} `)
	var schedSpecMin int = 0

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						Replicas: 1,
						Type:     arbv1.ResourceTypeNamespace,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericNamespaceAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "v1",
		"kind": "Namespace", 
	"metadata": {
		"name": "aw-generic-namespace-0",
		"labels": {
			"app": "aw-generic-namespace-0"
		}
	}} `)
	var schedSpecMin int = 0

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						DesiredAvailable: 0,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createStatefulSetAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "StatefulSet", 
	"metadata": {
		"name": "aw-statefulset-2",
		"namespace": "test",
		"labels": {
			"app": "aw-statefulset-2"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-statefulset-2"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-statefulset-2"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-statefulset-2"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-statefulset-2",
						"image": "k8s.gcr.io/echoserver:1.4",
						"imagePullPolicy": "Never",
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypeStatefulSet,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericStatefulSetAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "apps/v1",
		"kind": "StatefulSet", 
	"metadata": {
		"name": "aw-generic-statefulset-2",
		"namespace": "test",
		"labels": {
			"app": "aw-generic-statefulset-2"
		}
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "aw-generic-statefulset-2"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "aw-generic-statefulset-2"
				},
				"annotations": {
					"appwrapper.mcad.ibm.com/appwrapper-name": "aw-generic-statefulset-2"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "aw-generic-statefulset-2",
						"image": "k8s.gcr.io/echoserver:1.4",
						"imagePullPolicy": "Never",
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item1"),
							Namespace: context.namespace,
						},
						DesiredAvailable: 2,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}
	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

//NOTE: Recommend this test not to be the last test in the test suite it may pass
//      may pass the local test but may cause controller to fail which is not
//      part of this test's validation.
func createBadPodTemplateAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {
			"labels": {
				"app": "aw-bad-podtemplate-2"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "aw-bad-podtemplate-2"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "aw-bad-podtemplate-2",
					"image": "k8s.gcr.io/echoserver:1.4",
					"ports": [
						{
							"containerPort": 80
						}
					]
				}
			]
		}
	} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item"),
							Namespace: context.namespace,
						},
						Replicas: 2,
						Type:     arbv1.ResourceTypePod,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createPodTemplateAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"metadata": 
	{
		"name": "aw-podtemplate-2",
		"namespace": "test",
		"labels": {
			"app": "aw-podtemplate-2"
		}
	},
	"template": {
		"metadata": {
			"labels": {
				"app": "aw-podtemplate-2"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "aw-podtemplate-2"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "aw-podtemplate-2",
					"image": "k8s.gcr.io/echoserver:1.4",
					"ports": [
						{
							"containerPort": 80
						}
					]
				}
			]
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item"),
							Namespace: context.namespace,
						},
						Replicas: 2,
						Type:     arbv1.ResourceTypePod,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createPodCheckFailedStatusAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"metadata": 
	{
		"name": "aw-checkfailedstatus-1",
		"namespace": "test",
		"labels": {
			"app": "aw-checkfailedstatus-1"
		}
	},
	"template": {
		"metadata": {
			"labels": {
				"app": "aw-checkfailedstatus-1"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "aw-checkfailedstatus-1"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "aw-checkfailedstatus-1",
					"image": "k8s.gcr.io/echoserver:1.4",
					"ports": [
						{
							"containerPort": 80
						}
					]
				}
			],
			"tolerations": [
				{
					"effect": "NoSchedule",
					"key": "key1",
					"operator": "Exists"
				}
			]
		}
	}} `)
	var schedSpecMin int = 1

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				Items: []arbv1.AppWrapperResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item"),
							Namespace: context.namespace,
						},
						Replicas: 1,
						Type:     arbv1.ResourceTypePod,
						Template: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericPodAWCustomDemand(context *context, name string, cpuDemand string) *arbv1.AppWrapper {

	genericItems := fmt.Sprintf(`{
		"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {
			"name": "%s",
			"namespace": "test",
			"labels": {
				"app": "%s"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "%s"
			}
		},
		"spec": {
			"containers": [
					{
						"name": "%s",
						"image": "k8s.gcr.io/echoserver:1.4",
						"resources": {
							"limits": {
								"cpu": "%s"
							},
							"requests": {
								"cpu": "%s"
							}
						},
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
			]
		}
	} `, name, name, name, name, cpuDemand, cpuDemand)

	rb := []byte(genericItems)
	var schedSpecMin int = 1

	labels := make(map[string]string)
	labels["quota_service"] = "service-w"

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
			Labels:    labels,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item"),
							Namespace: context.namespace,
						},
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericPodAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {
			"name": "aw-generic-pod-1",
			"namespace": "test",
			"labels": {
				"app": "aw-generic-pod-1"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "aw-generic-pod-1"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "aw-generic-pod-1",
					"image": "k8s.gcr.io/echoserver:1.4",
					"resources": {
						"limits": {
							"memory": "150Mi"
						},
						"requests": {
							"memory": "150Mi"
						}
					},
					"ports": [
						{
							"containerPort": 80
						}
					]
				}
			]
		}
	} `)

	var schedSpecMin int = 1

	labels := make(map[string]string)
	labels["quota_service"] = "service-w"

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
			Labels:    labels,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item"),
							Namespace: context.namespace,
						},
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createGenericPodTooBigAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {
			"name": "aw-generic-big-pod-1",
			"namespace": "test",
			"labels": {
				"app": "aw-generic-big-pod-1"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "aw-generic-big-pod-1"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "aw-generic-big-pod-1",
					"image": "k8s.gcr.io/echoserver:1.4",
					"resources": {
						"limits": {
							"cpu": "100",
							"memory": "150Mi"
						},
						"requests": {
							"cpu": "100",
							"memory": "150Mi"
						}
					},
					"ports": [
						{
							"containerPort": 80
						}
					]
				}
			]
		}
	} `)

	var schedSpecMin int = 1

	labels := make(map[string]string)
	labels["quota_service"] = "service-w"

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
			Labels:    labels,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item"),
							Namespace: context.namespace,
						},
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}

func createBadGenericPodAW(context *context, name string) *arbv1.AppWrapper {
	rb := []byte(`{"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {
			"labels": {
				"app": "aw-bad-generic-pod-1"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "aw-bad-generic-pod-1"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "aw-bad-generic-pod-1",
					"image": "k8s.gcr.io/echoserver:1.4",
					"ports": [
						{
							"containerPort": 80
						}
					]
				}
			]
		}
	} `)
	var schedSpecMin int = 1

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item"),
							Namespace: context.namespace,
						},
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).NotTo(HaveOccurred())

	return appwrapper
}
func createBadGenericPodTemplateAW(context *context, name string) (*arbv1.AppWrapper, error) {
	rb := []byte(`{"metadata": 
	{
		"name": "aw-generic-podtemplate-2",
		"namespace": "test",
		"labels": {
			"app": "aw-generic-podtemplate-2"
		}
	},
	"template": {
		"metadata": {
			"labels": {
				"app": "aw-generic-podtemplate-2"
			},
			"annotations": {
				"appwrapper.mcad.ibm.com/appwrapper-name": "aw-generic-podtemplate-2"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "aw-generic-podtemplate-2",
					"image": "k8s.gcr.io/echoserver:1.4",
					"ports": [
						{
							"containerPort": 80
						}
					]
				}
			]
		}
	}} `)
	var schedSpecMin int = 2

	aw := &arbv1.AppWrapper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
		},
		Spec: arbv1.AppWrapperSpec{
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: schedSpecMin,
			},
			AggrResources: arbv1.AppWrapperResourceList{
				GenericItems: []arbv1.AppWrapperGenericResource{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("%s-%s", name, "item"),
							Namespace: context.namespace,
						},
						DesiredAvailable: 2,
						GenericTemplate: runtime.RawExtension{
							Raw: rb,
						},
					},
				},
			},
		},
	}

	appwrapper, err := context.karclient.ArbV1().AppWrappers(context.namespace).Create(aw)
	Expect(err).To(HaveOccurred())
	return appwrapper, err
}

func deleteReplicaSet(ctx *context, name string) error {
	foreground := metav1.DeletePropagationForeground
	return ctx.kubeclient.AppsV1().ReplicaSets(ctx.namespace).Delete(gcontext.Background(), name, metav1.DeleteOptions{
		PropagationPolicy: &foreground,
	})
}

func deleteAppWrapper(ctx *context, name string) error {
	foreground := metav1.DeletePropagationForeground
	return ctx.karclient.ArbV1().AppWrappers(ctx.namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &foreground,
	})
}

func replicaSetReady(ctx *context, name string) wait.ConditionFunc {
	return func() (bool, error) {
		deployment, err := ctx.kubeclient.ExtensionsV1beta1().ReplicaSets(ctx.namespace).Get(gcontext.Background(), name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		pods, err := ctx.kubeclient.CoreV1().Pods(ctx.namespace).List(gcontext.Background(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		labelSelector := labels.SelectorFromSet(deployment.Spec.Selector.MatchLabels)

		readyTaskNum := 0
		for _, pod := range pods.Items {
			if !labelSelector.Matches(labels.Set(pod.Labels)) {
				continue
			}
			if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded {
				readyTaskNum++
			}
		}

		return *(deployment.Spec.Replicas) == int32(readyTaskNum), nil
	}
}

func waitReplicaSetReady(ctx *context, name string) error {
	return wait.Poll(100*time.Millisecond, ninetySeconds, replicaSetReady(ctx, name))
}

func clusterSize(ctx *context, req v1.ResourceList) int32 {
	nodes, err := ctx.kubeclient.CoreV1().Nodes().List(gcontext.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	pods, err := ctx.kubeclient.CoreV1().Pods(metav1.NamespaceAll).List(gcontext.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	used := map[string]*csapi.Resource{}

	for _, pod := range pods.Items {
		nodeName := pod.Spec.NodeName
		if len(nodeName) == 0 || pod.DeletionTimestamp != nil {
			continue
		}

		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			continue
		}

		if _, found := used[nodeName]; !found {
			used[nodeName] = csapi.EmptyResource()
		}

		for _, c := range pod.Spec.Containers {
			req := csapi.NewResource(c.Resources.Requests)
			used[nodeName].Add(req)
		}
	}

	res := int32(0)

	for _, node := range nodes.Items {
		// Skip node with taints
		if len(node.Spec.Taints) != 0 {
			continue
		}

		alloc := csapi.NewResource(node.Status.Allocatable)
		slot := csapi.NewResource(req)

		// Removed used resources.
		if res, found := used[node.Name]; found {
			_, err := alloc.Sub(res)
			Expect(err).NotTo(HaveOccurred())
		}

		for slot.LessEqual(alloc) {
			_, err := alloc.Sub(slot)
			Expect(err).NotTo(HaveOccurred())
			res++
		}
	}

	return res
}

func clusterNodeNumber(ctx *context) int {
	nodes, err := ctx.kubeclient.CoreV1().Nodes().List(gcontext.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	nn := 0
	for _, node := range nodes.Items {
		if len(node.Spec.Taints) != 0 {
			continue
		}
		nn++
	}

	return nn
}

func computeNode(ctx *context, req v1.ResourceList) (string, int32) {
	nodes, err := ctx.kubeclient.CoreV1().Nodes().List(gcontext.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	pods, err := ctx.kubeclient.CoreV1().Pods(metav1.NamespaceAll).List(gcontext.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	used := map[string]*csapi.Resource{}

	for _, pod := range pods.Items {
		nodeName := pod.Spec.NodeName
		if len(nodeName) == 0 || pod.DeletionTimestamp != nil {
			continue
		}

		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			continue
		}

		if _, found := used[nodeName]; !found {
			used[nodeName] = csapi.EmptyResource()
		}

		for _, c := range pod.Spec.Containers {
			req := csapi.NewResource(c.Resources.Requests)
			used[nodeName].Add(req)
		}
	}

	for _, node := range nodes.Items {
		if len(node.Spec.Taints) != 0 {
			continue
		}

		res := int32(0)

		alloc := csapi.NewResource(node.Status.Allocatable)
		slot := csapi.NewResource(req)

		// Removed used resources.
		if res, found := used[node.Name]; found {
			_, err := alloc.Sub(res)
			Expect(err).NotTo(HaveOccurred())
		}

		for slot.LessEqual(alloc) {
			_, err := alloc.Sub(slot)
			Expect(err).NotTo(HaveOccurred())
			res++
		}

		if res > 0 {
			return node.Name, res
		}
	}

	return "", 0
}

func getPodsOfAppWrapper(ctx *context, aw *arbv1.AppWrapper) []*v1.Pod {
	aw, err := ctx.karclient.ArbV1().AppWrappers(aw.Namespace).Get(aw.Name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	pods, err := ctx.kubeclient.CoreV1().Pods(aw.Namespace).List(gcontext.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	var awpods []*v1.Pod

	for index, _ := range pods.Items {
		// Get a pointer to the pod in the list not a pointer to the podCopy
		pod := &pods.Items[index]

		if gn, found := pod.Annotations[arbv1.AppWrapperAnnotationKey]; !found || gn != aw.Name {
			continue
		}
		awpods = append(awpods, pod)
	}

	return awpods
}

func taintAllNodes(ctx *context, taints []v1.Taint) error {
	nodes, err := ctx.kubeclient.CoreV1().Nodes().List(gcontext.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	for _, node := range nodes.Items {
		newNode := node.DeepCopy()

		newTaints := newNode.Spec.Taints
		for _, t := range taints {
			found := false
			for _, nt := range newTaints {
				if nt.Key == t.Key {
					found = true
					break
				}
			}

			if !found {
				newTaints = append(newTaints, t)
			}
		}

		newNode.Spec.Taints = newTaints

		patchBytes, err := preparePatchBytesforNode(node.Name, &node, newNode)
		Expect(err).NotTo(HaveOccurred())

		_, err = ctx.kubeclient.CoreV1().Nodes().Patch(gcontext.Background(), node.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		Expect(err).NotTo(HaveOccurred())
	}

	return nil
}

func removeTaintsFromAllNodes(ctx *context, taints []v1.Taint) error {
	nodes, err := ctx.kubeclient.CoreV1().Nodes().List(gcontext.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	for _, node := range nodes.Items {
		newNode := node.DeepCopy()

		var newTaints []v1.Taint
		for _, nt := range newTaints {
			found := false
			for _, t := range taints {
				if nt.Key == t.Key {
					found = true
					break
				}
			}

			if !found {
				newTaints = append(newTaints, nt)
			}
		}
		newNode.Spec.Taints = newTaints

		patchBytes, err := preparePatchBytesforNode(node.Name, &node, newNode)
		Expect(err).NotTo(HaveOccurred())

		_, err = ctx.kubeclient.CoreV1().Nodes().Patch(gcontext.Background(), node.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		Expect(err).NotTo(HaveOccurred())
	}

	return nil
}

func preparePatchBytesforNode(nodeName string, oldNode *v1.Node, newNode *v1.Node) ([]byte, error) {
	oldData, err := json.Marshal(oldNode)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal oldData for node %q: %v", nodeName, err)
	}

	newData, err := json.Marshal(newNode)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal newData for node %q: %v", nodeName, err)
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Node{})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateTwoWayMergePatch for node %q: %v", nodeName, err)
	}

	return patchBytes, nil
}
