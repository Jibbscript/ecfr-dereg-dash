import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import TitlePage from '../../pages/title/[t].vue'
import { defineComponent, Suspense, ref } from 'vue'

vi.mock('vue-router', () => ({
  useRoute: () => ({
    params: { t: '40' }
  })
}))

const UsaAccordion = { template: '<div><slot /></div>' }
const UsaAccordionItem = { 
    props: ['title'],
    template: '<div><h3>{{ title }}</h3><slot /></div>' 
}

const mockTitle = {
    total_words: 5000,
    avg_rscs: 12.5,
    summary: 'Title summary here.'
}

global.useFetch = vi.fn().mockResolvedValue({
    data: ref(mockTitle),
    error: ref(null)
})

describe('TitlePage', () => {
  it('renders title details', async () => {
    const TestComponent = defineComponent({
        components: { TitlePage },
        template: '<Suspense><TitlePage /></Suspense>'
    })

    const wrapper = mount(TestComponent, {
        global: {
            components: { UsaAccordion, UsaAccordionItem, TitlePage }
        }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Word Count: 5000')
    expect(wrapper.text()).toContain('RSCS: 12.5')
    expect(wrapper.text()).toContain('Title summary here')
  })

  it('logs error on failure', async () => {
       const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
       global.useFetch = vi.fn().mockResolvedValue({
            data: ref(null),
            error: ref('Fetch failed')
        })
        
        const TestComponent = defineComponent({
            components: { TitlePage },
            template: '<Suspense><TitlePage /></Suspense>'
        })

        mount(TestComponent, {
            global: { components: { UsaAccordion, UsaAccordionItem, TitlePage } }
        })
        
        await flushPromises()
        
        expect(consoleSpy).toHaveBeenCalledWith('Error loading title:', 'Fetch failed')
        consoleSpy.mockRestore()
  })
})
