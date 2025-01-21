export function updateImageLabel(store, hash, label) {
    store.update((images) =>
        images.map((img) =>
            img.filehash === hash
                ? { ...img, label, new_label: label, reviewed: true }
                : img
        )
    );
}

export function deleteImageFromStore(store, hash) {
    store.update((images) => images.filter((img) => img.filehash !== hash));
}

export function addNewUploads(store, newImages) {
    store.update((current) => {
        const existing = new Map(current.map((img) => [img.filehash, img]));

        newImages.forEach((img) => {
            if (existing.has(img.filehash)) {
                const existingImg = existing.get(img.filehash);
                existing.set(img.filehash, {
                    ...existingImg,
                    ...img,
                    reviewed: existingImg.reviewed || img.reviewed,
                    new_label: existingImg.new_label || img.new_label,
                });
            } else {
                existing.set(img.filehash, img);
            }
        });

        return Array.from(existing.values()).sort((a, b) => b.id - a.id);
    });
}