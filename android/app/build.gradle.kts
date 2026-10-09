plugins {
    alias(libs.plugins.androidApplication)
    alias(libs.plugins.kotlinAndroid)
    alias(libs.plugins.kotlinCompose)
}

android {
    namespace = "com.brackistar.gamemasternotes"
    compileSdk = libs.versions.compileSdk.get().toInt()

    defaultConfig {
        applicationId = "com.brackistar.gamemasternotes"
        minSdk = libs.versions.minSdk.get().toInt()
        targetSdk = libs.versions.targetSdk.get().toInt()
        versionCode = 1
        versionName = "0.1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        ndk {
            abiFilters += listOf("arm64-v8a")
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro",
            )
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    buildFeatures {
        compose = true
    }
}

kotlin {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}

dependencies {
    implementation(platform(libs.composeBom))
    implementation(project(":core:ai"))
    implementation(project(":core:data"))
    implementation(project(":core:design"))
    implementation(project(":core:domain"))
    implementation(project(":core:importpacks"))
    implementation(project(":core:retrieval"))
    implementation(project(":feature:assistant"))
    implementation(project(":feature:home"))
    implementation(project(":feature:import"))
    implementation(project(":feature:library"))
    implementation(project(":feature:settings"))

    implementation(libs.activityCompose)
    implementation(libs.composeMaterial3)
    implementation(libs.composeUi)
    implementation(libs.composeUiToolingPreview)
    implementation(libs.lifecycleRuntimeKtx)
    implementation(libs.lifecycleViewmodelCompose)
    implementation(libs.navigationCompose)
    implementation(libs.coroutinesAndroid)

    debugImplementation(libs.composeUiTooling)
    testImplementation(libs.junit)
}
