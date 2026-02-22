import type { Meta, StoryObj } from '@storybook/svelte';
import Select from './Select.svelte';
import '../../app.css';

const meta = {
  title: 'Components/Select',
  component: Select,
  tags: ['autodocs'],
  argTypes: {
    value: { control: 'text', description: 'Selected option value' },
    autofocus: { control: 'boolean' }
  }
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    label: 'Account Type',
    value: 'personal',
    options: [
      { value: 'personal', label: 'Personal' },
      { value: 'business', label: 'Business' },
      { value: 'enterprise', label: 'Enterprise' },
    ],
  },
};