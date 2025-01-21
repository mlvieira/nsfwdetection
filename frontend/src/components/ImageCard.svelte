<script>
    import Modal from "./Modal.svelte";

    export let upload;
    export let isPending = false;
    export let onLabel;
    export let onUpdate;
    export let onDelete;
    export let gridSize = "md";
    let isModalOpen = false;
    let image = upload;

    $: isEditMode = upload.reviewed;

    $: displayLabel = upload.reviewed
        ? upload.new_label || upload.label
        : upload.label;

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

<div class="relative bg-white shadow rounded overflow-hidden">
    <div
        class="absolute top-2 left-2 px-2 py-1 rounded bg-blue-500 text-white text-xs font-semibold uppercase shadow"
    >
        {isEditMode ? "Edit" : "Labeling"}
    </div>

    {#if isPending}
        <div
            class="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center"
        >
            <div
                class="loader border-4 border-blue-500 border-t-transparent rounded-full w-12 h-12 animate-spin"
            ></div>
        </div>
    {/if}

    <img
        src={upload.filepath}
        alt="Uploaded Image"
        class="w-full h-64 object-cover cursor-pointer"
        on:click={openModal}
    />

    <div class="p-4 text-center">
        <p class="text-gray-700 font-semibold">
            Confidence: {parseFloat(upload.confidence).toFixed(2)}%
        </p>
        <p class="text-sm text-gray-500 mb-2">{displayLabel || "Unlabeled"}</p>

        <div
            class="flex justify-center flex-wrap gap-2 mt-2 px-2 md:px-4 py-1 md:py-2"
        >
            <button
                class="rounded bg-green-500 text-white hover:bg-green-600 transition-all"
                class:px-2={gridSize === "lg"}
                class:px-3={gridSize === "md"}
                class:px-4={gridSize === "sm"}
                class:py-1={gridSize === "lg"}
                class:py-2={gridSize === "md"}
                class:text-xs={gridSize === "lg"}
                class:text-sm={gridSize === "md"}
                class:text-base={gridSize === "sm"}
                on:click={() => handleLabelOrUpdate("SFW")}
                disabled={isPending}
            >
                SFW
            </button>
            <button
                class="rounded bg-red-500 text-white hover:bg-red-600 transition-all"
                class:px-2={gridSize === "lg"}
                class:px-3={gridSize === "md"}
                class:px-4={gridSize === "sm"}
                class:py-1={gridSize === "lg"}
                class:py-2={gridSize === "md"}
                class:text-xs={gridSize === "lg"}
                class:text-sm={gridSize === "md"}
                class:text-base={gridSize === "sm"}
                on:click={() => handleLabelOrUpdate("NSFW")}
                disabled={isPending}
            >
                NSFW
            </button>
            <button
                class="rounded bg-gray-500 text-white hover:bg-gray-600 transition-all"
                class:px-2={gridSize === "lg"}
                class:px-3={gridSize === "md"}
                class:px-4={gridSize === "sm"}
                class:py-1={gridSize === "lg"}
                class:py-2={gridSize === "md"}
                class:text-xs={gridSize === "lg"}
                class:text-sm={gridSize === "md"}
                class:text-base={gridSize === "sm"}
                on:click={() => onDelete?.(upload.filehash)}
                disabled={isPending}
            >
                Delete
            </button>
        </div>
    </div>
</div>

<Modal bind:isOpen={isModalOpen} {image} onClose={closeModal} />
