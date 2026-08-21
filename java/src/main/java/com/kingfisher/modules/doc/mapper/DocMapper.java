package com.kingfisher.modules.doc.mapper;

import com.kingfisher.common.query.Condition;
import com.kingfisher.modules.doc.domain.DocDirectory;
import com.kingfisher.modules.doc.domain.DocVersion;
import com.kingfisher.modules.doc.domain.Document;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;

/**
 * 文档 Mapper，映射 SQLite doc_directories / documents / doc_versions 表。
 */
@Mapper
public interface DocMapper {

    // ---- 目录 ----
    List<DocDirectory> findAllDirs();

    DocDirectory findDirById(@Param("id") Long id);

    void insertDir(DocDirectory dir);

    void updateDir(@Param("id") Long id, @Param("updates") Map<String, Object> updates);

    void deleteDir(@Param("id") Long id);

    int countDirChildren(@Param("parentId") Long parentId);

    int countDirDocuments(@Param("dirId") Long dirId);

    // ---- 文档 ----
    List<Document> listDocs(@Param("dirId") Long dirId,
                            @Param("keyword") String keyword,
                            @Param("searchableColumns") List<String> searchableColumns,
                            @Param("filters") List<Condition> filters,
                            @Param("sort") String sort,
                            @Param("offset") int offset,
                            @Param("limit") int limit,
                            @Param("userId") Long userId,
                            @Param("isAdmin") boolean isAdmin);

    long countDocs(@Param("dirId") Long dirId,
                   @Param("keyword") String keyword,
                   @Param("searchableColumns") List<String> searchableColumns,
                   @Param("filters") List<Condition> filters,
                   @Param("userId") Long userId,
                   @Param("isAdmin") boolean isAdmin);

    List<Document> listAllVisibleDocs(@Param("userId") Long userId, @Param("isAdmin") boolean isAdmin);

    List<Document> listAllPublicDocs();

    Document getDocById(@Param("id") Long id, @Param("userId") Long userId, @Param("isAdmin") boolean isAdmin);

    Document getPublicDoc(@Param("id") Long id);

    void insertDoc(Document doc);

    void updateDoc(@Param("id") Long id, @Param("updates") Map<String, Object> updates);

    void setDocStatus(@Param("id") Long id, @Param("status") String status, @Param("publishedAt") LocalDateTime publishedAt);

    void deleteDoc(@Param("id") Long id);

    void deleteDocVersions(@Param("docId") Long docId);

    // ---- 版本 ----
    List<DocVersion> listVersions(@Param("docId") Long docId);

    DocVersion getVersion(@Param("docId") Long docId, @Param("versionNo") int versionNo);

    void insertVersion(DocVersion version);

    // ---- 辅助 ----
    Map<Long, String> loadUsernames();
}
