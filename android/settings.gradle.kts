pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "GameMasterNotes"
include(":app")
include(":core:ai")
include(":core:data")
include(":core:design")
include(":core:domain")
include(":core:importpacks")
include(":core:retrieval")
include(":feature:assistant")
include(":feature:home")
include(":feature:import")
include(":feature:library")
include(":feature:search")
include(":feature:session")
include(":feature:settings")
