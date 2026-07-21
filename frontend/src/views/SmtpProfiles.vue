<template>
  <section class="smtp-profiles">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">{{ $t('settings.smtp.name') }}</h1>
      </div>
      <div class="column has-text-right">
        <b-field expanded>
          <b-button expanded type="is-primary" icon-left="plus" class="btn-new" @click="showNewForm"
            :disabled="!$can('settings:manage')">
            {{ $t('globals.buttons.new') }}
          </b-button>
        </b-field>
      </div>
    </header>

    <b-table :data="profiles" :loading="loading" :hoverable="true">
      <b-table-column v-slot="props" field="name" :label="$t('globals.fields.name')" :td-attrs="$utils.tdID" sortable>
        <strong>{{ props.row.name }}</strong>
      </b-table-column>

      <b-table-column v-slot="props" field="host" label="SMTP Host" sortable>
        {{ props.row.host }}:{{ props.row.port }}
      </b-table-column>

      <b-table-column v-slot="props" field="username" label="Username" sortable>
        {{ props.row.username }}
      </b-table-column>

      <b-table-column v-slot="props" field="fromEmail" label="From Email" sortable>
        {{ props.row.fromEmail || '-' }}
      </b-table-column>

      <b-table-column v-slot="props" field="campaignCount" label="Campaigns" sortable>
        {{ props.row.campaignCount }}
      </b-table-column>

      <b-table-column v-slot="props" field="enabled" :label="$t('globals.fields.status')" sortable>
        <b-tag :type="props.row.enabled ? 'is-success' : 'is-danger'">
          {{ props.row.enabled ? $t('globals.buttons.enabled') : $t('globals.buttons.disabled') }}
        </b-tag>
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" align="right" :label="$t('globals.fields.actions')">
        <div>
          <b-tooltip :label="$t('globals.buttons.edit')" type="is-dark" position="is-bottom">
            <a href="#" @click.prevent="showEditForm(props.row)" class="mr-2" :disabled="!$can('settings:manage')" :aria-label="$t('globals.buttons.edit')">
              <b-icon icon="pencil-outline" />
            </a>
          </b-tooltip>

          <b-tooltip :label="$t('globals.buttons.duplicate')" type="is-dark" position="is-bottom">
            <a href="#" @click.prevent="duplicateProfile(props.row)" class="mr-2" :disabled="!$can('settings:manage')" :aria-label="$t('globals.buttons.duplicate')">
              <b-icon icon="content-copy" />
            </a>
          </b-tooltip>

          <b-tooltip :label="$t('settings.smtp.testConnection')" type="is-dark" position="is-bottom">
            <a href="#" @click.prevent="testProfile(props.row)" class="mr-2" :aria-label="$t('settings.smtp.testConnection')">
              <b-icon icon="email-outline" />
            </a>
          </b-tooltip>

          <b-tooltip v-if="props.row.campaignCount === 0" :label="$t('globals.buttons.delete')" type="is-dark"
            position="is-bottom">
            <a href="#" @click.prevent="$utils.confirm(null, () => deleteProfile(props.row))"
              :disabled="!$can('settings:manage')" :aria-label="$t('globals.buttons.delete')">
              <b-icon icon="trash-can-outline" />
            </a>
          </b-tooltip>
          <b-tooltip v-else :label="$t('settings.smtp.inUse')" type="is-dark" position="is-bottom">
            <span class="has-text-grey-light">
              <b-icon icon="trash-can-outline" />
            </span>
          </b-tooltip>
        </div>
      </b-table-column>
    </b-table>

    <!-- Create/Edit Modal -->
    <b-modal :active.sync="isModalActive" :width="640" scroll="keep">
      <form @submit.prevent="onSubmit">
        <div class="modal-card">
          <header class="modal-card-head">
            <p class="modal-card-title">{{ isEditing ? $t('globals.buttons.edit') : $t('globals.buttons.new') }}
              {{ $t('settings.smtp.name') }}</p>          </header>

          <section class="modal-card-body">
            <b-field :label="$t('settings.mailserver.profileName')" label-position="on-border" required>
              <b-input v-model="form.name" name="name" :maxlength="200" required />
            </b-field>

            <div class="columns">
              <div class="column is-9">
                <b-field :label="$t('settings.mailserver.host')" label-position="on-border" required>
                  <b-input v-model="form.host" name="host" placeholder="smtp.yourmailserver.net" :maxlength="200" required />
                </b-field>
              </div>
              <div class="column">
                <b-field :label="$t('settings.mailserver.port')" label-position="on-border" required>
                  <b-numberinput v-model="form.port" name="port" type="is-light" controls-position="compact"
                    placeholder="587" min="1" max="65535" />
                </b-field>
              </div>
            </div>

            <div class="columns">
              <div class="column">
                <b-field :label="$t('settings.mailserver.username')" label-position="on-border">
                  <b-input v-model="form.username" name="username" :maxlength="200" />
                </b-field>
              </div>
              <div class="column">
                <b-field :label="$t('settings.mailserver.password')" label-position="on-border">
                  <b-input v-model="form.password" name="password" type="password" :maxlength="200"
                    :placeholder="isEditing ? $t('settings.mailserver.passwordHelp') : ''" />
                </b-field>
              </div>
            </div>

            <div class="columns">
              <div class="column is-4">
                <b-field :label="$t('settings.mailserver.encryption')" label-position="on-border">
                  <b-select v-model="form.encryption" name="encryption" expanded>
                    <option value="starttls">STARTTLS</option>
                    <option value="ssl_tls">SSL/TLS</option>
                    <option value="none">{{ $t('settings.mailserver.none') }}</option>
                  </b-select>
                </b-field>
              </div>
            </div>

            <hr />

            <b-field :label="$t('campaigns.fromAddress')" label-position="on-border">
              <b-input v-model="form.fromEmail" name="from_email" placeholder="you@example.com" :maxlength="200" />
            </b-field>

            <div class="columns">
              <div class="column">
                <b-field :label="$t('settings.mailserver.fromName')" label-position="on-border">
                  <b-input v-model="form.fromName" name="from_name" :maxlength="200" />
                </b-field>
              </div>
              <div class="column">
                <b-field label="Reply-To" label-position="on-border">
                  <b-input v-model="form.replyTo" name="reply_to" placeholder="reply@example.com" :maxlength="200" />
                </b-field>
              </div>
            </div>

            <b-field>
              <b-switch v-model="form.enabled" name="enabled" :native-value="true">
                {{ $t('globals.buttons.enabled') }}
              </b-switch>
            </b-field>

            <div v-if="formTestResult" class="mt-2" :class="{ 'has-text-danger': formTestResult.status === 'error', 'has-text-success': formTestResult.status === 'success' }">
              <b-icon :icon="formTestResult.status === 'success' ? 'check-circle' : 'alert-circle'" />
              {{ formTestResult.message }}
            </div>
          </section>

          <footer class="modal-card-foot">
            <b-button @click="isModalActive = false">
              {{ $t('globals.buttons.cancel') }}
            </b-button>
            <b-button type="is-info" :loading="testing" @click="testCurrentForm" :disabled="!form.host">
              {{ $t('settings.smtp.testConnection') }}
            </b-button>
            <b-button native-type="submit" type="is-primary" :loading="saving">
              {{ $t('globals.buttons.save') }}
            </b-button>
          </footer>
        </div>
      </form>
    </b-modal>

    <!-- Test Connection Modal -->
    <b-modal :active.sync="isTestModalActive" :width="480" scroll="keep">
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ $t('settings.smtp.testConnection') }}</p>
        </header>
        <section class="modal-card-body">
          <b-field :message="$t('campaigns.sendTestHelp')">
            <b-input v-model="testEmail" :placeholder="$t('campaigns.testEmails')" type="email" />
          </b-field>
          <b-button type="is-primary" :loading="testing" @click="sendTest">
            {{ $t('settings.smtp.sendTest') }}
          </b-button>
          <div v-if="testResult" class="mt-3" :class="{ 'has-text-danger': testResult.status === 'error', 'has-text-success': testResult.status === 'success' }">
            <b-icon :icon="testResult.status === 'success' ? 'check-circle' : 'alert-circle'" />
            {{ testResult.message }}
          </div>
        </section>
        <footer class="modal-card-foot">
          <b-button @click="isTestModalActive = false">
            {{ $t('globals.buttons.close') }}
          </b-button>
        </footer>
      </div>
    </b-modal>
  </section>
</template>

<script>
import Vue from 'vue';
import {
  getSMTPProfiles,
  createSMTPProfile,
  updateSMTPProfile,
  deleteSMTPProfile,
  duplicateSMTPProfile,
  testSMTPProfile,
} from '../api';

export default Vue.extend({
  data() {
    return {
      profiles: [],
      loading: false,
      saving: false,
      isModalActive: false,
      isEditing: false,
      editId: null,
      form: this.getEmptyForm(),

      isTestModalActive: false,
      testProfileData: null,
      testEmail: '',
      testing: false,
      testResult: null,
      formTestResult: null,
    };
  },

  methods: {
    getEmptyForm() {
      return {
        name: '',
        host: '',
        port: 587,
        username: '',
        password: '',
        encryption: 'starttls',
        fromEmail: '',
        fromName: '',
        replyTo: '',
        enabled: true,
      };
    },

    async loadProfiles() {
      this.loading = true;
      try {
        const data = await getSMTPProfiles();
        this.profiles = data;
      } finally {
        this.loading = false;
      }
    },

    showNewForm() {
      this.isEditing = false;
      this.editId = null;
      this.form = this.getEmptyForm();
      this.formTestResult = null;
      this.isModalActive = true;
    },

    showEditForm(profile) {
      this.isEditing = true;
      this.editId = profile.id;
      this.form = {
        name: profile.name,
        host: profile.host,
        port: profile.port,
        username: profile.username,
        password: '',
        encryption: profile.encryption,
        fromEmail: profile.fromEmail,
        fromName: profile.fromName,
        replyTo: profile.replyTo,
        enabled: profile.enabled,
      };
      this.formTestResult = null;
      this.isModalActive = true;
    },

    async onSubmit() {
      this.saving = true;
      try {
        if (this.isEditing) {
          await updateSMTPProfile(this.editId, this.form);
        } else {
          await createSMTPProfile(this.form);
        }
        this.isModalActive = false;
        this.loadProfiles();
      } finally {
        this.saving = false;
      }
    },

    async deleteProfile(profile) {
      try {
        await deleteSMTPProfile(profile.id);
        this.loadProfiles();
      } catch (e) {
        // Error toast is shown by the API interceptor.
      }
    },

    async duplicateProfile(profile) {
      try {
        await duplicateSMTPProfile(profile.id);
        this.loadProfiles();
      } catch (e) {
        // Error toast is shown by the API interceptor.
      }
    },

    testProfile(profile) {
      this.testProfileData = profile;
      this.testEmail = '';
      this.testResult = null;
      this.isTestModalActive = true;
    },

    async sendTest() {
      this.testing = true;
      this.testResult = null;
      try {
        const payload = {
          ...this.testProfileData,
          email: this.testEmail,
        };
        const data = await testSMTPProfile(payload);
        this.testResult = data;
      } catch (e) {
        this.testResult = { status: 'error', message: e.toString() };
      } finally {
        this.testing = false;
      }
    },

    async testCurrentForm() {
      this.testing = true;
      this.formTestResult = null;
      try {
        const data = await testSMTPProfile({
          name: this.form.name,
          host: this.form.host,
          port: this.form.port,
          username: this.form.username,
          password: this.form.password,
          encryption: this.form.encryption,
          from_email: this.form.fromEmail,
          from_name: this.form.fromName,
          reply_to: this.form.replyTo,
          email: '',
        });
        this.formTestResult = data;
      } catch (e) {
        this.formTestResult = { status: 'error', message: e.toString() };
      } finally {
        this.testing = false;
      }
    },
  },

  created() {
    this.loadProfiles();
  },
});
</script>
