plugins {
    alias(libs.plugins.androidLibrary)
    alias(libs.plugins.kotlinAndroid)
    alias(libs.plugins.ksp)
}

android {
    namespace = "com.brackistar.gamemasternotes.core.data"
    compileSdk = libs.versions.compileSdk.get().toInt()

    defaultConfig {
        minSdk = libs.versions.minSdk.get().toInt()
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    testOptions {
        unitTests.isIncludeAndroidResources = true
    }
}

kotlin {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}

ksp {
    arg("room.generateKotlin", "true")
}

dependencies {
    implementation(project(":core:domain"))
    api(project(":core:retrieval"))
    implementation(libs.roomRuntime)
    implementation(libs.roomKtx)
    implementation(libs.coroutinesAndroid)
    ksp(libs.roomCompiler)
    testImplementation(libs.roomTesting)
    testImplementation(libs.androidxTestCore)
    testImplementation(libs.junit)
    testImplementation(libs.coroutinesTest)
    testImplementation(libs.robolectric)
}
