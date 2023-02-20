package toolbox

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/cluster"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
)

func (self *toolboxImpl) newKubernetesParser() *cobra.Command {
	root := &cobra.Command{
		Use:   "k8s",
		Short: "Kubernetes interactive command",
		Long: "Interact with kubernetes cloud cluster throught the toolbox " +
			"and perform SRE job script for dedicated intend",
	}

	root.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Get list of clusters",
		Run: self.GenerateSafeCallback(
			"k8s-status",
			func(cmd *cobra.Command, args []string) {
				clusterMgr, err := container.Pick("cluster")
				if err != nil {
					self.Fail("Get cluster got error: %v", err)
					return
				}

				tenant, err := cluster.Pick(clusterMgr, args[0])
				if err != nil {
					self.Fail("Pick %s got error: %v", args[0], err)
					return
				}

				client, err := tenant.GetClient()
				if err != nil {
					self.Fail("Get client got error: %v", err)
					return
				}

				pods, err := client.GetPods("")
				if err != nil {
					self.Fail("Fail get pods: %v", err)
				}

				for _, pod := range pods.Items {
					if pod.Status.Phase != "Running" {
						self.Fail("Fail pod %s", pod.ObjectMeta.Name)
						return
					}
				}

				self.Ok("Seem Ok now!!!")
			},
		),
	})
	root.AddCommand(self.newKubernetesGetParser())
	return root
}

func (self *toolboxImpl) newKubernetesGetParser() *cobra.Command {
	root := &cobra.Command{
		Use:   "get",
		Short: "Get some resource from kubernetes",
	}

	getClustersCmd := &cobra.Command{
		Use:   "clusters",
		Short: "Get list of clusters",
		Run: self.GenerateSafeCallback(
			"k8s-get-pods",
			func(cmd *cobra.Command, args []string) {
				clusterMgr, err := container.Pick("cluster")
				if err != nil {
					self.Fail("Get cluster got error: %v", err)
					return
				}

				clusters, err := cluster.List(clusterMgr)
				if err != nil {
					self.Fail("Fetch clusters fails: %v", err)
					return
				}

				self.Ok("%v", clusters)
			},
		),
	}

	getAppPodsCmd := &cobra.Command{
		Use:   "app",
		Short: "Get list of application",
		Run: self.GenerateSafeCallback(
			"k8s-get-pods",
			func(cmd *cobra.Command, args []string) {
				ns, err := cmd.Flags().GetString("namespace")
				if err != nil {
					self.Fail("parse namespace fail: %v", err)
					return
				}

				clusterMgr, err := container.Pick("cluster")
				if err != nil {
					self.Fail("Get cluster got error: %v", err)
					return
				}

				tenant, err := cluster.Pick(clusterMgr, args[0])
				if err != nil {
					self.Fail("Pick %s got error: %v", args[0], err)
					return
				}

				client, err := tenant.GetClient()
				if err != nil {
					self.Fail("Get client got error: %v", err)
					return
				}

				pods, err := client.GetPods(ns)
				if err != nil {
					self.Fail("Fail get pods: %v", err)
				}

				metrics, err := client.GetPodMetrics()

				pvs, err := client.GetPVs()
				if err != nil {
					self.Fail("Fail get pods: %v", err)
				}

				mapClaimToPV := make(map[string]corev1.PersistentVolume)
				for _, pv := range pvs.Items {
					mapClaimToPV[pv.Spec.ClaimRef.Name] = pv
				}

				mapPodToMetric := make(map[string][]kubernetes.Container)

				for _, item := range metrics.Items {
					if len(ns) > 0 && ns != item.Metadata.Namespace {
						continue
					}

					mapPodToMetric[item.Metadata.Name] = item.Containers
				}

				mapAppToPods := make(map[string][]corev1.Pod)

				for _, pod := range pods.Items {
					appName, foundApp := pod.Labels["app"]
					_, foundHash := pod.Labels["pod-template-hash"]

					if !foundApp || !foundHash || len(pod.Labels) > 2 {
						self.Fail("found wrong app %s", appName)
						continue
					}

					if _, existed := mapAppToPods[appName]; !existed {
						mapAppToPods[appName] = make([]corev1.Pod, 0)
					}

					mapAppToPods[appName] = append(mapAppToPods[appName], pod)
				}

				cnt := 0
				for name, pods := range mapAppToPods {
					self.Ok("- App %s:", name)
					for _, pod := range pods {
						self.Ok("- Pod %s:  %s", pod.ObjectMeta.Name, pod.Status.Phase)

						for _, container := range pod.Spec.Containers {
							self.Ok("    - Container %s:%s", container.Name, container.Image)
						}
					}

					self.Ok("")

					cnt += 1
					if cnt == 5 {
						self.Flush()
						cnt = 0
					}
				}
			},
		),
	}
	getAppPodsCmd.PersistentFlags().
		String("namespace", "default", "The k8s namespace we would like to access")

	getInfraPodsCmd := &cobra.Command{
		Use:   "infra",
		Short: "Get list of infrastructure pods",
		Run: self.GenerateSafeCallback(
			"k8s-get-infra",
			func(cmd *cobra.Command, args []string) {
				ns, err := cmd.Flags().GetString("namespace")
				if err != nil {
					self.Fail("parse namespace fail: %v", err)
					return
				}

				clusterMgr, err := container.Pick("cluster")
				if err != nil {
					self.Fail("Get cluster got error: %v", err)
					return
				}

				tenant, err := cluster.Pick(clusterMgr, args[0])
				if err != nil {
					self.Fail("Pick %s got error: %v", args[0], err)
					return
				}

				client, err := tenant.GetClient()
				if err != nil {
					self.Fail("Get client got error: %v", err)
					return
				}

				pods, err := client.GetInfraPods(ns)
				if err != nil {
					self.Fail("Fail get pods: %v", err)
				}

				pvs, err := client.GetPVs()
				if err != nil {
					self.Fail("Fail get pods: %v", err)
				}

				mapClaimToPV := make(map[string]corev1.PersistentVolume)
				for _, pv := range pvs.Items {
					mapClaimToPV[pv.Spec.ClaimRef.Name] = pv
				}

				cnt := 0
				for _, pod := range pods.Items {
					ok := false

					for _, vol := range pod.Spec.Volumes {
						if vol.VolumeSource.PersistentVolumeClaim != nil {
							ok = true
							break
						}
					}

					if ok {
						self.Ok("- Pod %s:  %s", pod.ObjectMeta.Name, pod.Status.Phase)

						for _, vol := range pod.Spec.Volumes {
							if vol.PersistentVolumeClaim == nil {
								continue
							}

							pv, found := mapClaimToPV[vol.PersistentVolumeClaim.ClaimName]

							if found && pv.Spec.CSI != nil {
								self.Ok("  - Volume: %s link to %s", vol.Name, pv.Spec.CSI.VolumeHandle)
							}
						}

						cnt += 1
						if cnt == 10 {
							self.Flush()
							cnt = 0
						}
					}
				}
			},
		),
	}
	getInfraPodsCmd.PersistentFlags().
		String("namespace", "default", "The k8s namespace we would like to access")

	root.AddCommand(getClustersCmd)
	root.AddCommand(getAppPodsCmd)
	root.AddCommand(getInfraPodsCmd)
	return root
}
