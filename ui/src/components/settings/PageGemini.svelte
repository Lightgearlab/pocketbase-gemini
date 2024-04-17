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
    import SvelteMarkdown from "svelte-markdown";

    $pageTitle = "Gemini AI";

    let originalFormSettings = {};
    let formSettings = {};
    let formText = "";
    let resultText = "";
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
        isSaving = true;
        try {
            const res = await fetch("http://127.0.0.1:8090/ask", {
                method: "POST",
                body: JSON.stringify({
                    req: text,
                }),
            });
            const json = await res.json();
            var data = json["data"];
            resultText = data;
            //addSuccessToast(data);
        } catch (err) {
            ApiClient.error(err);
        }
        isSaving = false;
    }

    function init(settings = {}) {
        formSettings = {
            meta: settings?.meta.text || {},
        };

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
        <form class="panel" autocomplete="off">
            <div class="flex m-b-sm flex-gap-10">
                <span class="txt-xl">Text Prompt</span>
            </div>

            <div class="row" style="text-align: right;">
                <Field class="form-field required" name="meta.text" let:uniqueId>
                    <label for={uniqueId}>Text</label>
                    <input type="text" id={uniqueId} required bind:value={formText} />
                </Field>
                <button
                    type="submit"
                    class:btn-loading={isSaving}
                    disabled={isSaving}
                    class="btn btn-secondary"
                    on:click={() => save(formText)}
                >
                    <span class="txt">Send Question</span>
                </button>
            </div>
            <hr />
            <SvelteMarkdown source={resultText} />
        </form>
    </div>
</PageWrapper>
