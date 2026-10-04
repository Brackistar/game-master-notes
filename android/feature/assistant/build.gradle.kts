plugins {
    alias(libs.plugins.androidLibrary)
    alias(libs.plugins.kotlinAndroid)
    alias(libs.plugins.kotlinCompose)
}

android {
    namespace = "com.brackistar.gamemasternotes.feature.assistant"
    compileSdk = libs.versions.compileSdk.get().toInt()

    defaultConfig {
        minSdk = libs.versions.minSdk.get().toInt()
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
    implementation(project(":core:retrieval"))
    implementation(libs.activityCompose)
    implementation("androidx.compose.foundation:foundation")
    implementation(libs.composeMaterial3)
    implementation(libs.composeUi)
    implementation(libs.lifecycleRuntimeCompose)
    implementation(libs.lifecycleViewmodelCompose)
    testImplementation(libs.junit)
}
