package com.brackistar.gamemasternotes.core.importpacks

import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test

class PackSchemaVersionTest {
    @Test
    fun acceptsPackBuilderSchemaString() {
        assertEquals("1.0", validatePackSchemaVersion("1.0"))
    }

    @Test
    fun rejectsNumericSchemaVersion() {
        assertThrows(PackImportException::class.java) {
            validatePackSchemaVersion(1)
        }
    }

    @Test
    fun rejectsUnsupportedSchemaString() {
        assertThrows(PackImportException::class.java) {
            validatePackSchemaVersion("2.0")
        }
    }

    @Test
    fun failedFolderScanSkipsAbsenceBasedPruning() {
        assertEquals(false, shouldPruneAfterScan(listOf("broken.gmnpack: invalid")))
        assertEquals(true, shouldPruneAfterScan(emptyList()))
    }
}
