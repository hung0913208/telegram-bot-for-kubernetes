package toolbox

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/application"
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

				appName := ""
				if len(args) > 1 {
					appName = args[1]
				}

				client, err := tenant.GetClient()
				if err != nil {
					self.Fail("Get client got error: %v", err)
					return
				}

				applicationMgr := application.NewApplicationManager(client)

				pods, err := applicationMgr.GetApplication(ns)
				if err != nil {
					self.Fail("Fail get pods: %v", err)
					return
				}

				metrics, err := client.GetPodMetrics()
				if err != nil {
					self.Fail("Fail get pods: %v", err)
					return
				}

				pvs, err := client.GetPVs()
				if err != nil {
					self.Fail("Fail get pods: %v", err)
					return
				}

				mapClaimToPV := make(map[string]corev1.PersistentVolume)
				for _, pv := range pvs.Items {
					mapClaimToPV[pv.Spec.ClaimRef.Name] = pv
				}

				mapMetricToPod := make(map[string][]kubernetes.Usage)
				for _, metric := range metrics.Items {
					for _, container := range metric.Containers {
						mapMetricToPod[metric.Metadata.Name] = append(
							mapMetricToPod[metric.Metadata.Name],
							container.Usage,
						)
					}
				}

				mapAppToPods := make(map[string][]corev1.Pod)
				for _, pod := range pods {
					appName, _ := pod.Labels["app"]

					if _, existed := mapAppToPods[appName]; !existed {
						mapAppToPods[appName] = make([]corev1.Pod, 0)
					}

					mapAppToPods[appName] = append(mapAppToPods[appName], pod)
				}

				cnt := 0
				for name, pods := range mapAppToPods {
					if len(appName) == 0 || name == appName {
						self.Ok("- App %s:", name)
						for _, pod := range pods {
							self.Ok("  - Pod %s:  %s", pod.ObjectMeta.Name, pod.Status.Phase)

							if len(appName) > 0 {
								for i, container := range pod.Spec.Containers {
									self.Ok("    - Container %s:%s", container.Name, container.Image)

									usages, ok := mapMetricToPod[pod.ObjectMeta.Name]
									if ok {
										self.Ok("       - CPU %s, Memory %s",
											usages[i].CPU,
											usages[i].Memory)
									}

									self.Ok("       - Image: %s", pod.Status.ContainerStatuses[i].ImageID)
									self.Ok("       - Restart: %d", pod.Status.ContainerStatuses[i].RestartCount)
								}
							}
						}

						self.Ok("")

						cnt += 1
						if cnt == 20 {
							self.Flush()
							cnt = 0
						}
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

				pods, err := client.GetHelmPods(ns)
				if err != nil {
					self.Fail("Fail get pods: %v", err)
					return
				}

				metrics, err := client.GetPodMetrics()
				if err != nil {
					self.Fail("Fail get metrics: %v", err)
					return
				}

				pvs, err := client.GetPVs()
				if err != nil {
					self.Fail("Fail get pods: %v", err)
					return
				}

				mapClaimToPV := make(map[string]corev1.PersistentVolume)
				for _, pv := range pvs.Items {
					mapClaimToPV[pv.Spec.ClaimRef.Name] = pv
				}

				mapMetricToPod := make(map[string][]kubernetes.Usage)
				for _, metric := range metrics.Items {
					for _, container := range metric.Containers {
						mapMetricToPod[metric.Metadata.Name] = append(
							mapMetricToPod[metric.Metadata.Name],
							container.Usage,
						)
					}
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

						usages, ok := mapMetricToPod[pod.ObjectMeta.Name]
						if ok {
							for i, usage := range usages {
								self.Ok("  - Container #%d: CPU %s, Memory %s",
									i+1,
									usage.CPU,
									usage.Memory)
							}
						}

						cnt += 1
						if cnt == 5 {
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
