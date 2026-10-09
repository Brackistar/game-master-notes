package com.brackistar.gamemasternotes.core.importpacks

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder

class VectorSidecarStoreTest {
    @get:Rule val temporaryFolder = TemporaryFolder()

    @Test fun installsChecksAndPrunesSidecars() {
        val directory = temporaryFolder.newFolder("vectors")
        val store = VectorSidecarStore(directory)
        store.installAtomic(
            packId = "pack/a",
            fingerprint = "fingerprint-1",
            modelId = "model",
            modelRevision = "revision",
            dimensions = 2,
            chunkIds = listOf("a", "b"),
            npyBytes = byteArrayOf(1, 2, 3),
        )

        assertTrue(store.isCurrent("pack/a", "fingerprint-1", "model", "revision"))
        assertFalse(store.isCurrent("pack/a", "fingerprint-2", "model", "revision"))

        store.pruneToPackIds(emptySet())
        assertFalse(store.isCurrent("pack/a", "fingerprint-1", "model", "revision"))
    }
}
