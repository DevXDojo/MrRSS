import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it } from 'vitest';
import { PhCircle, PhLinkSimple } from '@phosphor-icons/vue';
import ContextMenu from './ContextMenu.vue';

afterEach(() => {
  document.body.innerHTML = '';
});

describe('ContextMenu', () => {
  it('renders built-in menu icons and legacy names without a global icon plugin', () => {
    const iconNames = [
      'PhArrowBendLeftDown',
      'PhArrowBendRightUp',
      'PhArrowSquareOut',
      'PhArrowsClockwise',
      'PhArticle',
      'PhBinoculars',
      'PhCheckCircle',
      'PhCircle',
      'PhClockCountdown',
      'PhCopy',
      'PhDownloadSimple',
      'PhEnvelope',
      'PhEnvelopeOpen',
      'PhEye',
      'PhEyeSlash',
      'PhFolderMinus',
      'PhGlobe',
      'PhLinkSimple',
      'PhMagnifyingGlass',
      'PhMagnifyingGlassPlus',
      'PhPencil',
      'PhPushPin',
      'PhRssSimple',
      'PhStar',
      'PhTextT',
      'PhTrash',
      'ph-link',
      'ph-link-simple',
      'ph-text-t',
      'ph-check-circle',
      'ph-circle',
      'ph-globe',
      'ph-pencil',
      'ph-trash',
      'ph-envelope',
      'ph-envelope-open',
      'ph-star',
      'ph-article',
      'ph-eye',
      'ph-eye-slash',
      'ph-arrow-square-out',
      'ph-clock-countdown',
      'ph-arrow-bend-right-up',
      'ph-arrow-bend-left-down',
      'ph-rss-simple',
    ];
    const wrapper = mount(ContextMenu, {
      props: {
        items: iconNames.map((icon) => ({ label: icon, icon })),
        x: 10,
        y: 10,
      },
    });

    try {
      expect(wrapper.findAll('svg')).toHaveLength(iconNames.length);
      expect(wrapper.findAllComponents(PhLinkSimple)).toHaveLength(3);
    } finally {
      wrapper.unmount();
    }
  });

  it('keeps menu labels when an icon is missing or unsupported', () => {
    const iconNames = [
      'PhUnknownIcon',
      'obsidian',
      'notion',
      'zotero',
      'toString',
      'constructor',
      '__proto__',
    ];
    const wrapper = mount(ContextMenu, {
      props: {
        items: [...iconNames.map((icon) => ({ label: icon, icon })), { label: 'No icon' }],
        x: 10,
        y: 10,
      },
    });

    try {
      expect(wrapper.findAll('svg')).toHaveLength(0);
      expect(wrapper.findAll('span').map((label) => label.text())).toEqual([
        ...iconNames,
        'No icon',
      ]);
    } finally {
      wrapper.unmount();
    }
  });

  it('preserves icon weights, sizes and menu colors', () => {
    const weights = ['regular', 'bold', 'light', 'fill', 'duotone', 'thin'] as const;
    const wrapper = mount(ContextMenu, {
      props: {
        items: [
          ...weights.map((iconWeight) => ({
            label: iconWeight,
            icon: 'ph-circle',
            iconWeight,
            iconColor: 'text-yellow-500',
          })),
          { label: 'Danger', icon: 'ph-link', danger: true },
        ],
        x: 10,
        y: 10,
      },
    });

    try {
      const circles = wrapper.findAllComponents(PhCircle);
      expect(circles.map((icon) => icon.props('weight'))).toEqual(weights);
      for (const icon of circles) {
        expect(icon.props('size')).toBe(20);
        expect(icon.classes()).toContain('text-yellow-500');
      }
      const dangerIcon = wrapper.getComponent(PhLinkSimple);
      expect(dangerIcon.props('weight')).toBe('regular');
      expect(dangerIcon.classes()).toEqual(
        expect.arrayContaining(['text-red-600', 'dark:text-red-400'])
      );
    } finally {
      wrapper.unmount();
    }
  });

  it('closes when an overlay stops bubbling outside clicks', async () => {
    const wrapper = mount(ContextMenu, {
      attachTo: document.body,
      props: {
        items: [{ label: 'Action', action: 'action' }],
        x: 10,
        y: 10,
      },
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));

    const overlay = document.createElement('div');
    overlay.addEventListener('click', (event) => event.stopPropagation());
    document.body.appendChild(overlay);
    overlay.dispatchEvent(new MouseEvent('click', { bubbles: true }));

    expect(wrapper.emitted('close')).toHaveLength(1);
    wrapper.unmount();
  });
});
