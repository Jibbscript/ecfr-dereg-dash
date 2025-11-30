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

    // Word count now displayed inside a summary box, allow formatted number
    expect(wrapper.text()).toMatch(/5,?000/) // matches 5000 or 5,000
    // RSCS value displayed with label "Avg. RSCS Score"
    expect(wrapper.text()).toContain('Avg. RSCS')
    expect(wrapper.text()).toContain('12.5')
    // Summary text appears in accordion section
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
