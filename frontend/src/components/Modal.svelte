<script>
    export let isOpen = false;
    export let image = null;
    export let onClose;

    function handleKeyDown(event) {
        if (event.key === "Escape") {
            onClose();
        }
    }

    $: if (isOpen) {
        document.body.style.overflow = "hidden";
        window.addEventListener("keydown", handleKeyDown);
    } else {
        document.body.style.overflow = "";
        window.removeEventListener("keydown", handleKeyDown);
    }

    function closeModal() {
        onClose();
    }
</script>

{#if isOpen}
    <div
        class="fixed inset-0 flex items-center justify-center bg-black bg-opacity-75 z-50 p-4"
        on:click={onClose}
    >
        <div
            class="relative bg-white rounded shadow-lg max-w-3xl w-full max-h-[90vh] overflow-auto p-4"
            on:click|stopPropagation
        >
            {#if image}
                <img
                    src={image.filepath}
                    alt="Expanded Image"
                    class="rounded object-contain max-h-[80vh] w-full"
                />
            {:else}
                <p class="text-center text-gray-500">No image available</p>
            {/if}

            <button
                class="absolute top-2 right-2 w-8 h-8 flex items-center justify-center rounded-full bg-gray-200 hover:bg-gray-300 text-gray-600 hover:text-gray-900 text-2xl"
                on:click={onClose}
            >
                &times;
            </button>
        </div>
    </div>
{/if}
