import { useState } from 'react'
import { useLaunchContext } from '../../hooks/useLaunchContext'
import {
  DOC_GROUPS,
  findDocById,
  getDefaultDoc,
  getAdjacentDocs,
  renderDocWithLaunchContext,
} from './catalog'
import { LaunchDocsFlowPage } from './LaunchDocsFlowPage'
import { LaunchDocsLayout } from './LaunchDocsLayout'
import { LaunchDocsMarkdownContent } from './LaunchDocsMarkdownContent'
import { LaunchDocsPager } from './LaunchDocsPager'
import { LaunchDocsSidebar } from './LaunchDocsSidebar'
import { StructuredApiDocsPage } from './StructuredApiDocsPage'
import {
  getStructuredApiParentDocId,
  isStructuredApiDocId,
  isStructuredApiEndpointDocId,
  type StructuredApiDocId,
} from './structuredApiDocs'

interface LaunchDocsPanelProps {
  embedded?: boolean
  initialDocId?: string
}

export function LaunchDocsPanel({ embedded = false, initialDocId }: LaunchDocsPanelProps) {
  const firstDoc = getDefaultDoc()
  const [activeId, setActiveId] = useState(() => {
    if (initialDocId && findDocById(initialDocId)) {
      return initialDocId
    }
    return firstDoc.id
  })
  const { launchBaseUrl, apiAuth } = useLaunchContext()

  const activeDoc = findDocById(activeId) || firstDoc
  const { previous, next } = isStructuredApiEndpointDocId(activeDoc.id)
    ? { previous: null, next: null }
    : getAdjacentDocs(activeDoc.id)
  const sidebarActiveId = isStructuredApiDocId(activeDoc.id)
    ? getStructuredApiParentDocId(activeDoc.id)
    : activeDoc.id

  const selectDoc = (id: string) => {
    const doc = findDocById(id)
    if (!doc) {
      return
    }
    setActiveId(doc.id)
  }

  const renderedContent = renderDocWithLaunchContext(activeDoc.content, launchBaseUrl, apiAuth.header)

  return (
    <LaunchDocsLayout
      embedded={embedded}
      sidebar={(
        <LaunchDocsSidebar
          groups={DOC_GROUPS}
          activeId={sidebarActiveId}
          onSelect={selectDoc}
        />
      )}
      header={null}
      content={(
        <div className="space-y-5">
          {activeDoc.id === 'tutorial-flow'
            ? <LaunchDocsFlowPage baseUrl={launchBaseUrl} />
            : isStructuredApiDocId(activeDoc.id)
              ? (
                <StructuredApiDocsPage
                  docId={activeDoc.id as StructuredApiDocId}
                  launchBaseUrl={launchBaseUrl}
                  authHeader={apiAuth.header}
                  onOpenDoc={selectDoc}
                />
              )
              : <LaunchDocsMarkdownContent content={renderedContent} docId={activeDoc.id} />}
          <LaunchDocsPager
            previous={previous}
            next={next}
            onSelect={selectDoc}
          />
        </div>
      )}
    />
  )
}
