<script>
    import { onMount } from "svelte";
    import { showToast } from "../utils/toast";
    import { stats } from "../services/api";
    import { token } from "../stores/auth";

    let isLoading = false;
    let averageConfidence = 0;
    let labelDistribution = {};
    let labelingEfficiencyPercentage = 0;
    let reviewedImages = 0;
    let totalImages = 0;
    let unlabeledImages = 0;

    async function loadStats() {
        if (isLoading) return;
        isLoading = true;

        try {
            const response = await stats($token);
            averageConfidence = response.average_confidence;
            labelDistribution = response.label_distribution;
            labelingEfficiencyPercentage =
                response.labeling_efficiency_percentage;
            reviewedImages = response.reviewed_images;
            totalImages = response.total_images;
            unlabeledImages = response.unlabeled_images;
        } catch (err) {
            showToast(err.message || "Failed to fetch stats", "error");
        } finally {
            isLoading = false;
        }
    }

    onMount(() => {
        loadStats();
    });
</script>

<div class="p-4">
    <div
        class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-4"
    >
        <h1 class="text-2xl font-bold text-center sm:text-left">
            Image Analysis Statistics
        </h1>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <div class="bg-white p-4 rounded shadow">
            <h2 class="text-xl font-semibold">Average Confidence</h2>
            <p class="text-2xl text-gray-700">
                {averageConfidence.toFixed(2)}%
            </p>
        </div>

        <div class="bg-white p-4 rounded shadow">
            <h2 class="text-xl font-semibold">Label Distribution</h2>
            <ul class="list-disc pl-5">
                {#each Object.entries(labelDistribution) as [label, count]}
                    <li class="text-lg text-gray-700">{label}: {count}</li>
                {/each}
            </ul>
        </div>

        <div class="bg-white p-4 rounded shadow">
            <h2 class="text-xl font-semibold">Labeling Efficiency</h2>
            <p class="text-2xl text-gray-700">
                {labelingEfficiencyPercentage.toFixed(2)}%
            </p>
        </div>

        <div class="bg-white p-4 rounded shadow">
            <h2 class="text-xl font-semibold">Reviewed Images</h2>
            <p class="text-2xl text-gray-700">
                {reviewedImages} / {totalImages}
            </p>
        </div>

        <div class="bg-white p-4 rounded shadow">
            <h2 class="text-xl font-semibold">Unlabeled Images</h2>
            <p class="text-2xl text-gray-700">{unlabeledImages}</p>
        </div>
    </div>
</div>
