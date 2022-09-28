package greenplumcluster

import (
	"context"
	"fmt"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const NodeCountErrorFmt = "there must be at least two nodes available to both master and segments for anti-affinity: the number of nodes available for master is %d and for segment is %d"
const AntiAffinityMismatchErrorFmt = "master and segment antiAffinity must be the same value: segment antiAffinity is %s, and master antiAffinity is %s"

func handleAntiAffinity(ctx context.Context, c client.Client, greenplumCluster greenplumv1.GreenplumCluster) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("antiAffinity: %w", err)
		}
	}()

	if greenplumCluster.Spec.MasterAndStandby.AntiAffinity == "yes" ||
		greenplumCluster.Spec.Segments.AntiAffinity == "yes" {

		var masterNodeList corev1.NodeList
		masterWorkerSelectorLabel := greenplumCluster.Spec.MasterAndStandby.WorkerSelector
		err = c.List(ctx, &masterNodeList, client.MatchingLabels(masterWorkerSelectorLabel))
		if err != nil {
			return fmt.Errorf("master node worker selector list: %w", err)
		}

		var segmentNodeList corev1.NodeList
		segmentWorkerSelectorLabel := greenplumCluster.Spec.Segments.WorkerSelector
		err = c.List(ctx, &segmentNodeList, client.MatchingLabels(segmentWorkerSelectorLabel))
		if err != nil {
			return fmt.Errorf("segment node worker selector list: %w", err)
		}

		// Look at nodes for each anti-affinity set and make sure there are at least two nodes of each type (masterx2 and segmentx2)
		valid, err := isAntiAffinityValid(greenplumCluster, masterNodeList, segmentNodeList)
		if !valid {
			return fmt.Errorf("instance %s does not meet requirements: %w", greenplumCluster.Name, err)
		}

		// Label every other node for master/standby respectively
		masterNodeLabelKey := fmt.Sprintf("greenplum-affinity-%s-master", greenplumCluster.Namespace)
		err = labelAlternateNodes(ctx, c, masterNodeList, masterNodeLabelKey, "true", "true")
		if err != nil {
			return err
		}

		// Label every other node for primary/mirror, respectively
		segNodeLabelKey := fmt.Sprintf("greenplum-affinity-%s-segment", greenplumCluster.Namespace)
		err = labelAlternateNodes(ctx, c, segmentNodeList, segNodeLabelKey, "a", "b")
		if err != nil {
			return err
		}
	}

	return nil
}

func isAntiAffinityValid(greenplumCluster greenplumv1.GreenplumCluster, masterNodeList, segmentNodeList corev1.NodeList) (bool, error) {
	masterAntiAffinity := greenplumCluster.Spec.MasterAndStandby.AntiAffinity
	segAntiAffinity := greenplumCluster.Spec.Segments.AntiAffinity
	if masterAntiAffinity != segAntiAffinity {
		return false, fmt.Errorf(AntiAffinityMismatchErrorFmt, segAntiAffinity, masterAntiAffinity)
	}

	numMasterNodes := len(masterNodeList.Items)
	numSegmentNodes := len(segmentNodeList.Items)
	if numMasterNodes < 2 || numSegmentNodes < 2 {
		return false, fmt.Errorf(NodeCountErrorFmt, numMasterNodes, numSegmentNodes)
	}

	return true, nil
}

func labelAlternateNodes(ctx context.Context, c client.Client, nodeList corev1.NodeList, key, evenVal, oddVal string) error {
	for i, node := range nodeList.Items {
		v := oddVal
		if i%2 == 0 {
			v = evenVal
		}
		labeledNode := node.DeepCopy()
		if labeledNode.Labels == nil {
			labeledNode.Labels = make(map[string]string)
		}
		labeledNode.Labels[key] = v
		if err := c.Patch(ctx, labeledNode, client.MergeFrom(&node)); err != nil {
			return fmt.Errorf("failed to add label '%s=%s' to node '%s': %w", key, v, node.Name, err)
		}
	}

	return nil
}
