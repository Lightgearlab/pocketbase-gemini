<script>
    import { slide } from "svelte/transition";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { pageTitle } from "@/stores/app";
    import { removeError } from "@/stores/errors";
    import { addSuccessToast } from "@/stores/toasts";
    import tooltip from "@/actions/tooltip";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import Field from "@/components/base/Field.svelte";
    import Toggler from "@/components/base/Toggler.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import SettingsSidebar from "@/components/settings/SettingsSidebar.svelte";
    import BackupsList from "@/components/settings/BackupsList.svelte";
    import S3Fields from "@/components/settings/S3Fields.svelte";
    import BackupUploadButton from "@/components/settings/BackupUploadButton.svelte";

    $pageTitle = "Gemini";

    let originalFormSettings = {};
    let formSettings = {};
    let isLoading = false;
    let isSaving = false;
    let initialHash = "";

    $: initialHash = JSON.stringify(originalFormSettings);

    $: hasChanges = initialHash != JSON.stringify(formSettings);

    loadSettings();

    async function loadSettings() {
        isLoading = true;

        try {
            const settings = (await ApiClient.settings.getAll()) || {};
            init(settings);
        } catch (err) {
            ApiClient.error(err);
        }

        isLoading = false;
    }

    async function save(text) {
        if (isSaving || !hasChanges) {
            return;
        }

        isSaving = true;

        try {
            const res = await fetch("http://localhost:8090/ask", {
                method: "POST",
                body: JSON.stringify({
                    req: text,
                }),
            });

            const json = await res.json();
            result = JSON.stringify(json);

            addSuccessToast(result);
        } catch (err) {
            ApiClient.error(err);
        }

        isSaving = false;
    }

    function init(settings = {}) {
        formSettings = {};

        originalFormSettings = JSON.parse(JSON.stringify(formSettings));
    }
</script>

<SettingsSidebar />

<PageWrapper>
    <header class="page-header">
        <nav class="breadcrumbs">
            <div class="breadcrumb-item">Settings</div>
            <div class="breadcrumb-item">{$pageTitle}</div>
        </nav>
    </header>

    <div class="wrapper">
        <div class="panel" autocomplete="off" on:submit|preventDefault={save}>
            <div class="flex m-b-sm flex-gap-10">
                <span class="txt-xl">Gemini AI</span>
            </div>

            <div class="col-lg-6">
                <Field class="form-field required" name="meta.text" let:uniqueId>
                    <label for={uniqueId}>Text</label>
                    <input type="text" id={uniqueId} required bind:value={formSettings.meta.text} />
                </Field>
            </div>
            <hr />
            <button type="submit" class="btn btn-expanded" on:click={() => save(formSettings.meta.text)}>
                <span class="txt">Send Question</span>
            </button>
            <!-- <div class="grid">
                <div class="col-lg-6">
                    <Field class="form-field required" name="meta.apikey" let:uniqueId>
                        <label for={uniqueId}>API KEY</label>
                        <input type="text" id={uniqueId} required bind:value={formSettings.meta.apikey} />
                    </Field>
                </div>
            </div> -->

            <!-- <button
                type="button"
                class="btn btn-secondary"
                class:btn-loading={isLoading}
                disabled={isLoading}
                on:click={() => (showBackupsSettings = !showBackupsSettings)}
            >
                <span class="txt">Backups options</span>
                {#if showBackupsSettings}
                    <i class="ri-arrow-up-s-line" />
                {:else}
                    <i class="ri-arrow-down-s-line" />
                {/if}
            </button> -->
        </div>
    </div>
</PageWrapper>
