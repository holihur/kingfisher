package com.kingfisher.modules.message.service;

import com.kingfisher.common.query.Query;
import com.kingfisher.modules.message.domain.Message;
import com.kingfisher.modules.message.mapper.MessageMapper;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
public class MessageService {

    private final MessageMapper mapper;
    public MessageService(MessageMapper mapper) { this.mapper = mapper; }

    public List<Message> listByRecipient(Long recipientId, Query q) {
        return mapper.listByRecipient(recipientId, q.getKeyword(), q.searchableColumns(), q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
    }
    public long countByRecipient(Long recipientId, Query q) {
        return mapper.countByRecipient(recipientId, q.getKeyword(), q.searchableColumns(), q.getFilters());
    }
    public Message getById(Long id, Long recipientId) { return mapper.findByIdAndRecipient(id, recipientId); }

    public Message create(Long senderId, String senderType, Long recipientId, String title, String content) {
        Message m = new Message();
        m.setSenderId(senderId); m.setSenderType(senderType); m.setRecipientId(recipientId);
        m.setTitle(title); m.setContent(content); m.setStatus("sent"); m.setIsRead(false);
        mapper.insert(m);
        return m;
    }

    public int sendBatch(Long senderId, String senderType, List<Long> recipientIds, String title, String content) {
        long batchId = System.nanoTime();
        for (Long rid : recipientIds) {
            Message m = new Message();
            m.setSenderId(senderId); m.setSenderType(senderType); m.setRecipientId(rid); m.setBatchId(batchId);
            m.setTitle(title); m.setContent(content); m.setStatus("sent"); m.setIsRead(false);
            mapper.insert(m);
        }
        return recipientIds.size();
    }

    public void markRead(Long id, Long recipientId) { mapper.markRead(id, recipientId, LocalDateTime.now()); }
    public void deleteBatch(List<Long> ids, Long recipientId) { mapper.deleteBatch(ids, recipientId); }
    public long unreadCount(Long recipientId) { return mapper.countUnread(recipientId); }
    public void revoke(Long id, Long senderId) { mapper.revoke(id, senderId); }
    public void revokeBatch(Long batchId, Long senderId) { mapper.revokeBatch(batchId, senderId); }
    public List<Message> listBatchMessages(Long batchId, Long senderId) { return mapper.listBatchMessages(batchId, senderId); }
    public List<Message> listSent(Long senderId, Query q) {
        return mapper.listBySender(senderId, q.getKeyword(), q.searchableColumns(), q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
    }
    public long countSent(Long senderId, Query q) { return mapper.countBySender(senderId, q.getKeyword(), q.searchableColumns(), q.getFilters()); }
}
