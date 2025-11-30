import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import IndexPage from '../../pages/index.vue'

describe('IndexPage', () => {
  let fetchMock: any

  beforeEach(() => {
    fetchMock = vi.fn()
    global.fetch = fetchMock
  })

  it('renders loading state initially', () => {
    fetchMock.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve([])
    })
    
    const wrapper = mount(IndexPage, {
      global: {
        stubs: {
          LoadingSkeleton: { template: '<div role="status">Loading…</div>' }
        }
      }
    })
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
  })

  it('renders data after fetch', async () => {
    const mockData = [
      { id: 1, name: 'Dept A', total_words: 1000, avg_rscs: 10, parent_id: null, lsa_counts: 5 },
      { id: 2, name: 'Agency B', total_words: 500, avg_rscs: 5, parent_id: 1, lsa_counts: 2 }
    ]

    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockData)
    })

    const wrapper = mount(IndexPage)
    await flushPromises()

    expect(wrapper.text()).not.toContain('Loading agency data...')
    expect(wrapper.text()).toContain('Dept A')
    expect(wrapper.text()).toContain('1,000')
  })

  it('handles fetch error', async () => {
    fetchMock.mockRejectedValue(new Error('Network error'))

    const wrapper = mount(IndexPage)
    await flushPromises()

    // Error is displayed via UsaAlert component
    expect(wrapper.text()).toContain('Network error')
  })

  it('filters by title', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([])
    })

    const wrapper = mount(IndexPage)
    await flushPromises()

    const select = wrapper.find('#title-filter')
    await select.setValue('10')
    
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('title=10'))
  })

  it('toggles checksum', async () => {
     fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([])
    })

    const wrapper = mount(IndexPage)
    await flushPromises()

    const checkbox = wrapper.find('#include-checksum')
    await checkbox.setValue(true)

    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('include_checksum=true'))
  })

  it('sorts data', async () => {
    const mockData = [
        { id: 1, name: 'B Dept', total_words: 100, avg_rscs: 10 },
        { id: 2, name: 'A Dept', total_words: 200, avg_rscs: 20 }
    ]
     fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockData)
    })

    const wrapper = mount(IndexPage)
    await flushPromises()

    // Default sort by total_words desc -> A Dept (200) first
    let rows = wrapper.findAll('tbody > tr.parent-row')
    expect(rows[0].text()).toContain('A Dept') 

    // Click the sortable button inside the first header (Agency Name)
    await wrapper.findAll('thead th button').at(0)!.trigger('click')
    rows = wrapper.findAll('tbody > tr.parent-row')
    expect(rows[0].text()).toContain('B Dept')

    // Click again -> asc
    await wrapper.findAll('thead th button').at(0)!.trigger('click')
    rows = wrapper.findAll('tbody > tr.parent-row')
    expect(rows[0].text()).toContain('A Dept')
  })

  it('expands/collapses children', async () => {
      const mockData = [
      { id: 1, name: 'Dept A', total_words: 1000, avg_rscs: 10, parent_id: null },
      { id: 2, name: 'Agency B', total_words: 500, avg_rscs: 5, parent_id: 1 }
    ]

    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockData)
    })

    const wrapper = mount(IndexPage)
    await flushPromises()

    expect(wrapper.find('.child-row').exists()).toBe(false)

    await wrapper.find('.parent-row').trigger('click')
    expect(wrapper.find('.child-row').exists()).toBe(true)
    expect(wrapper.find('.child-row').text()).toContain('Agency B')
  })

  it('computes summary stats correctly', async () => {
      const mockData = [
      { id: 1, name: 'Dept A', total_words: 1000, avg_rscs: 10, parent_id: null },
      { id: 2, name: 'Dept B', total_words: 2000, avg_rscs: 20, parent_id: null }
    ]

    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockData)
    })

    const wrapper = mount(IndexPage, {
      global: {
        stubs: {
          // Stub MetricCard to render props so text assertions work reliably
          MetricCard: {
            props: ['title','value','format','description','hasInfo'],
            template: '<div class="metric-card-stub"><span class="title">{{ title }}</span> <span class="value">{{ value }}</span></div>'
          }
        }
      }
    })
    await flushPromises()

expect(/3,?000/.test(wrapper.text())).toBe(true)
    expect(wrapper.text()).toContain('15')
  })

  it('displays checksum when enabled', async () => {
      const mockData = [
      { id: 1, name: 'Dept A', total_words: 1000, avg_rscs: 10, parent_id: null, content_checksum: 'abcdef123456' }
    ]
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockData)
    })

    const wrapper = mount(IndexPage)
    await flushPromises()

    const checkbox = wrapper.find('#include-checksum')
    await checkbox.setValue(true)
    
    expect(wrapper.text()).toContain('abcdef12...')
  })

  it('handles empty rscs for average', async () => {
      const mockData = [
      { id: 1, name: 'Dept A', total_words: 1000, avg_rscs: 0, parent_id: null }
    ]
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockData)
    })
    
    const wrapper = mount(IndexPage)
    await flushPromises()
    
    // RSCS of 0 should be shown in MetricCard
    expect(wrapper.text()).toContain('0.0')
  })
})
