import { ref, computed } from 'vue';
import type { AIProfile, AIProfileTestResult, AIProfileFormData } from '@/types/aiProfile';
import { defaultAIProfileFormData } from '@/types/aiProfile';

// Shared state for AI profiles
const profiles = ref<AIProfile[]>([]);
const isLoading = ref(false);
const error = ref<string | null>(null);

export function useAIProfiles() {
  // Computed properties
  const defaultProfile = computed(
    () => profiles.value.find((p) => p.is_default) || profiles.value[0]
  );

  const hasProfiles = computed(() => profiles.value.length > 0);

  // Fetch all profiles
  async function fetchProfiles(): Promise<void> {
    isLoading.value = true;
    error.value = null;

    try {
      const response = await fetch('/api/ai/profiles');
      if (!response.ok) {
        throw new Error(`Failed to fetch profiles: ${response.status}`);
      }
      profiles.value = await response.json();
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch profiles';
      console.error('Error fetching AI profiles:', e);
    } finally {
      isLoading.value = false;
    }
  }

  // Get a single profile by ID (with full details including masked API key)
  async function getProfile(id: number): Promise<AIProfile | null> {
    try {
      const response = await fetch(`/api/ai/profiles/${id}`);
      if (!response.ok) {
        if (response.status === 404) return null;
        throw new Error(`Failed to get profile: ${response.status}`);
      }
      return await response.json();
    } catch (e) {
      console.error('Error getting AI profile:', e);
      return null;
    }
  }

  // Create a new profile
  async function createProfile(data: AIProfileFormData): Promise<AIProfile | null> {
    try {
      const response = await fetch('/api/ai/profiles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Failed to create profile: ${response.status}`);
      }

      const newProfile = await response.json();
      await fetchProfiles(); // Refresh the list
      return newProfile;
    } catch (e) {
      console.error('Error creating AI profile:', e);
      throw e;
    }
  }

  // Update an existing profile
  async function updateProfile(id: number, data: AIProfileFormData): Promise<AIProfile | null> {
    try {
      const response = await fetch(`/api/ai/profiles/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Failed to update profile: ${response.status}`);
      }

      const updatedProfile = await response.json();
      await fetchProfiles(); // Refresh the list
      return updatedProfile;
    } catch (e) {
      console.error('Error updating AI profile:', e);
      throw e;
    }
  }

  // Delete a profile
  async function deleteProfile(id: number): Promise<boolean> {
    try {
      const response = await fetch(`/api/ai/profiles/${id}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error(`Failed to delete profile: ${response.status}`);
      }

      await fetchProfiles(); // Refresh the list
      return true;
    } catch (e) {
      console.error('Error deleting AI profile:', e);
      return false;
    }
  }

  // Set a profile as default
  async function setDefaultProfile(id: number): Promise<boolean> {
    try {
      const response = await fetch(`/api/ai/profiles/${id}/default`, {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error(`Failed to set default profile: ${response.status}`);
      }

      await fetchProfiles(); // Refresh the list
      return true;
    } catch (e) {
      console.error('Error setting default AI profile:', e);
      return false;
    }
  }

  // Test a single profile
  async function testProfile(id: number): Promise<AIProfileTestResult | null> {
    try {
      const response = await fetch(`/api/ai/profiles/${id}/test`, {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error(`Failed to test profile: ${response.status}`);
      }

      return await response.json();
    } catch (e) {
      console.error('Error testing AI profile:', e);
      return null;
    }
  }

  // Test configuration without saving (for new profiles or unsaved changes)
  async function testConfig(config: {
    api_key: string;
    endpoint: string;
    model: string;
    custom_headers: string;
  }): Promise<AIProfileTestResult | null> {
    try {
      const response = await fetch('/api/ai/profiles/test-config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });

      if (!response.ok) {
        throw new Error(`Failed to test config: ${response.status}`);
      }

      return await response.json();
    } catch (e) {
      console.error('Error testing AI config:', e);
      return null;
    }
  }

  // Test all profiles
  async function testAllProfiles(): Promise<AIProfileTestResult[]> {
    try {
      const response = await fetch('/api/ai/profiles/test-all', {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error(`Failed to test profiles: ${response.status}`);
      }

      return await response.json();
    } catch (e) {
      console.error('Error testing AI profiles:', e);
      return [];
    }
  }

  // Get profile name by ID (useful for display)
  function getProfileName(id: number | string | null | undefined): string {
    if (!id) return '';
    const numId = typeof id === 'string' ? parseInt(id, 10) : id;
    const profile = profiles.value.find((p) => p.id === numId);
    return profile?.name || '';
  }

  // Create empty form data for new profile
  function createEmptyFormData(): AIProfileFormData {
    return { ...defaultAIProfileFormData };
  }

  // Convert profile to form data (for editing)
  function profileToFormData(profile: AIProfile): AIProfileFormData {
    return {
      name: profile.name,
      api_key: profile.api_key || '',
      endpoint: profile.endpoint,
      model: profile.model,
      custom_headers: profile.custom_headers,
      is_default: profile.is_default,
    };
  }

  return {
    // State
    profiles,
    isLoading,
    error,

    // Computed
    defaultProfile,
    hasProfiles,

    // Methods
    fetchProfiles,
    getProfile,
    createProfile,
    updateProfile,
    deleteProfile,
    setDefaultProfile,
    testProfile,
    testConfig,
    testAllProfiles,
    getProfileName,
    createEmptyFormData,
    profileToFormData,
  };
}
