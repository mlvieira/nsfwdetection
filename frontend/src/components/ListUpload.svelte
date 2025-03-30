<script>
    import Modal from "./Modal.svelte";

    export let upload;
    export let isPending = false;
    export let onLabel;
    export let onUpdate;
    export let onDelete;
    let isModalOpen = false;
    let image = upload;

    $: isEditMode = upload.reviewed;

    $: displayLabel = upload.reviewed
        ? upload.new_label
        : "Unlabeled";

    function handleLabelOrUpdate(label) {
        const handler = !isEditMode ? onLabel : onUpdate;
        handler?.(upload.filehash, label);
    }

    function openModal() {
        isModalOpen = true;
    }

    function closeModal() {
        isModalOpen = false;
    }
</script>

<div
    class="flex flex-col sm:flex-row items-start sm:items-center space-y-4 sm:space-y-0 sm:space-x-4 p-4 border rounded"
>
    <img
        src={upload.filepath}
        alt="Uploaded Image"
        class="w-20 h-20 object-cover rounded cursor-pointer"
        on:click={openModal}
    />

    <div class="flex-grow flex flex-col items-start space-y-2 w-full sm:w-auto">
        <div
            class="px-2 py-1 rounded bg-blue-500 text-white text-xs font-semibold uppercase shadow"
        >
            {isEditMode ? "Edit" : "Labeling"}
        </div>
        <p class="font-semibold text-gray-700">H: {displayLabel}</p>
        <p class="font-semibold text-gray-700">AI: {upload.label}</p>
        <p class="text-sm text-gray-500">
            Confidence: {parseFloat(upload.confidence).toFixed(2)}%
        </p>
    </div>

    <div class="flex flex-wrap justify-end gap-2 w-full sm:w-auto">
        <button
            class="px-3 py-1 rounded bg-green-500 text-white hover:bg-green-600"
            on:click={() => handleLabelOrUpdate("SFW")}
            disabled={isPending}
        >
            SFW
        </button>
        <button
            class="px-3 py-1 rounded bg-red-500 text-white hover:bg-red-600"
            on:click={() => handleLabelOrUpdate("NSFW")}
            disabled={isPending}
        >
            NSFW
        </button>
        <button
            class="px-3 py-1 rounded bg-gray-500 text-white hover:bg-gray-600"
            on:click={() => onDelete?.(upload.filehash)}
            disabled={isPending}
        >
            Delete
        </button>
    </div>
</div>

<Modal bind:isOpen={isModalOpen} {image} onClose={closeModal} />
