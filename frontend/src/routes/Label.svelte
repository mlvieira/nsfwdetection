<script>
  import { onMount } from "svelte";
  import {
    fetchUploads,
    labelImage,
    deleteImage,
    updateLabel,
  } from "../services/api";
  import { token } from "../stores/auth";
  import { uploads, newUploads } from "../stores/uploads";
  import { pendingRatings, completedRatings } from "../stores/ratings";
  import { debounce } from "../utils/debounce";
  import { showToast } from "../utils/toast";
  import { initWebSocket, closeWebSocket } from "../services/ws";
  import ImageCard from "../components/ImageCard.svelte";
  import ListItem from "../components/ListUpload.svelte";
  import {
    updateImageLabel,
    deleteImageFromStore,
    addNewUploads,
  } from "../utils/uploads";

  export let gridSize = "md";
  export let viewMode = "grid";
  let currentCursor = Number.MAX_SAFE_INTEGER;
  let limit = 20;
  let totalCount = 0;
  let loadedCount = 0;
  let isLoading = false;
  let allLoaded = false;
  let reviewed;
  let unsubscribeNewUploads;

  $: gridSize = localStorage.getItem("gridSize") || "md";
  $: if (gridSize) localStorage.setItem("gridSize", gridSize);

  $: viewMode = localStorage.getItem("viewMode") || "grid";
  $: if (viewMode) localStorage.setItem("viewMode", viewMode);

  async function loadUploads() {
    if (isLoading || allLoaded) return;
    isLoading = true;

    try {
      const response = await fetchUploads(
        currentCursor,
        limit,
        reviewed,
        $token,
      );

      addNewUploads(uploads, response.data);
      loadedCount += response.count;
      totalCount = response.total;
      allLoaded = loadedCount >= totalCount;

      if (response.data.length > 0) {
        currentCursor = response.data[response.data.length - 1].id;
      }
    } catch (err) {
      showToast(err.message || "Failed to fetch uploads", "error");
    } finally {
      isLoading = false;
    }
  }

  async function handleLabelImage(hash, label) {
    try {
      pendingRatings.update((set) => set.add(hash));
      await labelImage(hash, label, $token);
      updateImageLabel(uploads, hash, label);
      completedRatings.update((set) => set.add(hash));
      showToast("Image labeled successfully", "success");
    } catch (err) {
      showToast(err.message || "Failed to label image", "error");
    } finally {
      pendingRatings.update((set) => {
        set.delete(hash);
        return set;
      });
    }
  }

  async function handleUpdateLabel(hash, newLabel) {
    try {
      await updateLabel(hash, newLabel, $token);
      updateImageLabel(uploads, hash, newLabel);
      completedRatings.update((set) => set.add(hash));
      showToast("Label updated successfully", "success");
    } catch (err) {
      showToast(err.message || "Failed to update label", "error");
    }
  }

  async function handleDeleteImage(hash) {
    try {
      await deleteImage(hash, $token);
      deleteImageFromStore(uploads, hash);
      showToast("Image deleted successfully", "success");
    } catch (err) {
      showToast(err.message || "Failed to delete image", "error");
    }
  }

  const debounceLoadMore = debounce(loadUploads, 300);

  onMount(() => {
    unsubscribeNewUploads = newUploads.subscribe((newImages) => {
      addNewUploads(uploads, newImages);
    });
    loadUploads();

    window.addEventListener("scroll", () => {
      if (
        window.innerHeight + window.scrollY >=
          document.body.offsetHeight - 200 &&
        !isLoading
      ) {
        debounceLoadMore();
      }
    });

    return () => {
      window.removeEventListener("scroll", debounceLoadMore);
      unsubscribeNewUploads?.();
    };
  });
</script>

<div class="p-4">
  <div
    class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-4"
  >
    <h1 class="text-2xl font-bold text-center sm:text-left">Label Images</h1>
    <div class="flex flex-wrap justify-center sm:justify-end gap-2">
      <button
        class="px-3 py-1 text-sm sm:px-4 sm:py-2 sm:text-base rounded transition-all duration-200"
        class:bg-blue-500={viewMode === "grid"}
        class:bg-gray-200={viewMode !== "grid"}
        class:text-white={viewMode === "grid"}
        class:text-gray-700={viewMode !== "grid"}
        class:hover:bg-blue-600={viewMode === "grid"}
        class:hover:bg-gray-300={viewMode !== "grid"}
        on:click={() => (viewMode = "grid")}
      >
        Grid View
      </button>
      <button
        class="px-3 py-1 text-sm sm:px-4 sm:py-2 sm:text-base rounded transition-all duration-200"
        class:bg-blue-500={viewMode === "list"}
        class:bg-gray-200={viewMode !== "list"}
        class:text-white={viewMode === "list"}
        class:text-gray-700={viewMode !== "list"}
        class:hover:bg-blue-600={viewMode === "list"}
        class:hover:bg-gray-300={viewMode !== "list"}
        on:click={() => (viewMode = "list")}
      >
        List View
      </button>
    </div>
    {#if viewMode === "grid"}
      <div class="flex flex-wrap justify-center sm:justify-end gap-2">
        <button
          class="px-3 py-1 text-sm sm:px-4 sm:py-2 sm:text-base rounded transition-all duration-200"
          class:bg-blue-500={gridSize === "sm"}
          class:bg-gray-200={gridSize !== "sm"}
          class:text-white={gridSize === "sm"}
          class:text-gray-700={gridSize !== "sm"}
          class:hover:bg-blue-600={gridSize === "sm"}
          class:hover:bg-gray-300={gridSize !== "sm"}
          on:click={() => (gridSize = "sm")}
        >
          Large
        </button>

        <button
          class="px-3 py-1 text-sm sm:px-4 sm:py-2 sm:text-base rounded transition-all duration-200"
          class:bg-blue-500={gridSize === "md"}
          class:bg-gray-200={gridSize !== "md"}
          class:text-white={gridSize === "md"}
          class:text-gray-700={gridSize !== "md"}
          class:hover:bg-blue-600={gridSize === "md"}
          class:hover:bg-gray-300={gridSize !== "md"}
          on:click={() => (gridSize = "md")}
        >
          Medium
        </button>

        <button
          class="px-3 py-1 text-sm sm:px-4 sm:py-2 sm:text-base rounded transition-all duration-200"
          class:bg-blue-500={gridSize === "lg"}
          class:bg-gray-200={gridSize !== "lg"}
          class:text-white={gridSize === "lg"}
          class:text-gray-700={gridSize !== "lg"}
          class:hover:bg-blue-600={gridSize === "lg"}
          class:hover:bg-gray-300={gridSize !== "lg"}
          on:click={() => (gridSize = "lg")}
        >
          Small
        </button>
      </div>
    {/if}
  </div>

  {#if isLoading && $uploads.length === 0}
    <div
      class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
    >
      {#each Array(10) as _, i}
        <div class="bg-gray-200 animate-pulse h-64 w-full rounded"></div>
      {/each}
    </div>
  {/if}

  {#if $uploads.length === 0 && !isLoading}
    <div class="text-center text-gray-500 py-10">No images to display</div>
  {:else if viewMode === "grid"}
    <div
      class={`grid gap-4 animate-fadeIn transition-all duration-300 ease-in-out ${
        gridSize === "sm"
          ? "grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6"
          : gridSize === "md"
            ? "grid-cols-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8"
            : "grid-cols-4 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10"
      }`}
    >
      {#each $uploads as upload}
        <ImageCard
          {upload}
          {gridSize}
          isPending={$pendingRatings.has(upload.filehash)}
          onLabel={handleLabelImage}
          onDelete={handleDeleteImage}
          onUpdate={handleUpdateLabel}
        />
      {/each}
    </div>
  {:else if viewMode === "list"}
    <div class="space-y-4">
      {#each $uploads as upload}
        <ListItem
          {upload}
          isPending={$pendingRatings.has(upload.filehash)}
          onLabel={handleLabelImage}
          onDelete={handleDeleteImage}
          onUpdate={handleUpdateLabel}
        />
      {/each}
    </div>
  {/if}

  <div class="text-center mt-6">
    <button
      on:click={debounceLoadMore}
      class="px-6 py-2 rounded bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 cursor-pointer"
      disabled={allLoaded}
    >
      Load More
    </button>
  </div>
</div>
