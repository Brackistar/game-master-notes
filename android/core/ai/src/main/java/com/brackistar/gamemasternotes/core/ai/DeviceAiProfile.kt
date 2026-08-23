package com.brackistar.gamemasternotes.core.ai

import android.app.ActivityManager
import android.content.Context
import android.os.Build

data class DeviceAiProfile(
    val totalRamMb: Long,
    val supportedAbis: List<String>,
    val isLowRamDevice: Boolean,
) {
    fun canRun(model: AiModel): Boolean =
        model.isFallback || (supportsArm64 && totalRamMb >= model.minimumRamMb)

    val supportsArm64: Boolean
        get() = supportedAbis.any { it == "arm64-v8a" }
}

object AndroidDeviceAiProfileReader {
    fun read(context: Context): DeviceAiProfile {
        val activityManager = context.getSystemService(Context.ACTIVITY_SERVICE) as ActivityManager
        val memoryInfo = ActivityManager.MemoryInfo()
        activityManager.getMemoryInfo(memoryInfo)
        return DeviceAiProfile(
            totalRamMb = memoryInfo.totalMem / BYTES_PER_MB,
            supportedAbis = Build.SUPPORTED_ABIS.toList(),
            isLowRamDevice = activityManager.isLowRamDevice,
        )
    }

    private const val BYTES_PER_MB = 1024L * 1024L
}
