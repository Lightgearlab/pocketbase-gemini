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
    let formSystemText = "";
    let formMaxTable = 4.0;
    let formMaxRow = 5.0;
    let resultText = "";
    let resultSystemText = "";
    let isLoading = false;
    let isSaving = false;
    let initialHash = "";
    let menuOpen = false;
    let inputValue = "";
    $: console.log(inputValue);

    const menuItems = ["Gemini", "ChatGPT"];
    let filteredItems = [];

    const handleInput = () => {
        return (filteredItems = menuItems.filter((item) =>
            item.toLowerCase().match(inputValue.toLowerCase()),
        ));
    };

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

    async function systemConfig(text, maxtable, maxrow) {
        isSaving = true;
        try {
            const res = await fetch("http://127.0.0.1:8090/gemini", {
                method: "POST",
                body: JSON.stringify({
                    req: text,
                    maxtable,
                    maxrow,
                }),
            });
            const json = await res.json();
            console.log(json);
            resultSystemText = json;
            addSuccessToast("created json config file config.json");
        } catch (err) {
            ApiClient.error(err);
        }
        isSaving = false;
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
            <div class="column">
                <div class="flex m-b-sm flex-gap-10">
                    <span class="txt-xl">JSON Generator</span>
                </div>
            </div>

            <div class="row" style="text-align: right;">
                <Field class="form-field required" name="meta.text" let:uniqueId>
                    <label for={uniqueId}>System</label>
                    <input type="text" id={uniqueId} required bind:value={formSystemText} />
                </Field>
                <Field class="form-field required" name="meta.maxtable" let:uniqueId>
                    <label for={uniqueId}>Max Table</label>
                    <input type="text" id={uniqueId} required bind:value={formMaxTable} />
                </Field>
                <Field class="form-field required" name="meta.maxrow" let:uniqueId>
                    <label for={uniqueId}>Max Row</label>
                    <input type="text" id={uniqueId} required bind:value={formMaxRow} />
                </Field>
                <button
                    type="submit"
                    class:btn-loading={isSaving}
                    disabled={isSaving}
                    class="btn btn-secondary"
                    on:click={() => systemConfig(formSystemText, formMaxTable, formMaxRow)}
                >
                    <span class="txt">Create</span>
                </button>
            </div>
            <hr />
            <Field class="form-field" name="collections" let:uniqueId>
                <label for={uniqueId} class="p-b-10">Collections</label>
                <textarea
                    id={uniqueId}
                    class="code"
                    spellcheck="false"
                    rows="15"
                    required
                    bind:value={resultSystemText}
                />

                <!-- {#if !!resultSystemText}
                    <div class="help-block help-block-error">Invalid collections configuration.</div>
                {/if} -->
            </Field>
        </form>
        <div class="m-20"></div>
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

<style>
    .dropdown {
        position: relative;
        display: inline-block;
    }

    .dropdown-content {
        display: none;
        position: absolute;
        background-color: #f6f6f6;
        min-width: 230px;
        border: 1px solid #ddd;
        z-index: 1;
    }

    /* Show the dropdown menu */
    .show {
        display: block;
    }
</style>
