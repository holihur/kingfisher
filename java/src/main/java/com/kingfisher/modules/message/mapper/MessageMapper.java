package com.kingfisher.modules.message.mapper;

import com.kingfisher.common.query.Condition;
import com.kingfisher.modules.message.domain.Message;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.time.LocalDateTime;
import java.util.List;

/**
 * 站内信 Mapper，映射 SQLite messages 表。
 */
@Mapper
public interface MessageMapper {

    /** 收件箱列表（按 recipient_id 过滤） */
    List<Message> listByRecipient(@Param("recipientId") Long recipientId,
                                  @Param("keyword") String keyword,
                                  @Param("searchableColumns") List<String> searchableColumns,
                                  @Param("filters") List<Condition> filters,
                                  @Param("sort") String sort,
                                  @Param("offset") int offset,
                                  @Param("limit") int limit);

    long countByRecipient(@Param("recipientId") Long recipientId,
                          @Param("keyword") String keyword,
                          @Param("searchableColumns") List<String> searchableColumns,
                          @Param("filters") List<Condition> filters);

    /** 管理端已发送列表（按 sender_id 过滤，去重取批次） */
    List<Message> listBySender(@Param("senderId") Long senderId,
                               @Param("keyword") String keyword,
                               @Param("searchableColumns") List<String> searchableColumns,
                               @Param("filters") List<Condition> filters,
                               @Param("sort") String sort,
                               @Param("offset") int offset,
                               @Param("limit") int limit);

    long countBySender(@Param("senderId") Long senderId,
                       @Param("keyword") String keyword,
                       @Param("searchableColumns") List<String> searchableColumns,
                       @Param("filters") List<Condition> filters);

    Message findByIdAndRecipient(@Param("id") Long id, @Param("recipientId") Long recipientId);

    void insert(Message message);

    void markRead(@Param("id") Long id, @Param("recipientId") Long recipientId, @Param("readAt") LocalDateTime readAt);

    void deleteBatch(@Param("ids") List<Long> ids, @Param("recipientId") Long recipientId);

    long countUnread(@Param("recipientId") Long recipientId);

    void revoke(@Param("id") Long id, @Param("senderId") Long senderId);

    void revokeBatch(@Param("batchId") Long batchId, @Param("senderId") Long senderId);

    List<Message> listBatchMessages(@Param("batchId") Long batchId, @Param("senderId") Long senderId);
}
