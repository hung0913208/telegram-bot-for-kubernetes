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

	getPodsCmd := &cobra.Command{
		Use:   "pods",
		Short: "Get list of pods",
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

				cnt := 0
				for _, pod := range pods.Items {
					self.Ok("- Pod %s:  %s", pod.ObjectMeta.Name, pod.Status.Phase)

					for _, vol := range pod.Spec.Volumes {
						if vol.PersistentVolumeClaim == nil {
							self.Ok("  - Volume: %s", vol.Name)
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
			},
		),
	}

	getPodsCmd.PersistentFlags().
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

	getAppPodsCmd := &cobra.Command{
		Use:   "app",
		Short: "Get list of application pods",
		Run: self.GenerateSafeCallback(
			"k8s-get-app",
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

				pods, err := client.GetAppPods(ns)
				if err != nil {
					self.Fail("Fail get pods: %v", err)
					return
				}

				cnt := 0
				for _, pod := range pods.Items {
					self.Ok("- Pod %s:  %s", pod.ObjectMeta.Name, pod.Status.Phase)
					cnt += 1
					if cnt == 10 {
						self.Flush()
						cnt = 0
					}
				}
			},
		),
	}
	getAppPodsCmd.PersistentFlags().
		String("namespace", "default", "The k8s namespace we would like to access")

	getPvcCmd := &cobra.Command{
		Use:   "pvc",
		Short: "Get list of application pods",
		Run: self.GenerateSafeCallback(
			"k8s-get-pvc-metric",
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

				nodes, err := client.GetNodes()
				if err != nil {
					self.Fail("Fail get nodes: %v", err)
					return
				}
				cnt := 0
				for _, node := range nodes.Items {
					metric, err := client.GetNodeMetrics(node.Name)
					if err != nil {
						self.Fail("Fail get metric of %s: %v", node.Name, err)
						continue
					}

					self.Ok("## Node %s:", node.Name)

					for _, pod := range metric.Pods {
						self.Ok("- Pod %s:", pod.PodRef.Name)

						for _, volume := range pod.Volumes {
							self.Ok(
								"  - Volume %s:\n"+
									"    - Used %dM\n"+
									"    - Capacity: %dM\n"+
									"    - Available: %dM",
								volume.Name,
								volume.UsedBytes/1048576,
								volume.CapacityBytes/1048576,
								volume.AvailableBytes/1048576,
							)
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

	root.AddCommand(getClustersCmd)
	root.AddCommand(getPodsCmd)
	root.AddCommand(getAppPodsCmd)
	root.AddCommand(getInfraPodsCmd)
	root.AddCommand(getPvcCmd)
	return root
}
