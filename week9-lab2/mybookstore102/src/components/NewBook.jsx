import React, { useState, useEffect } from 'react';
import BookCard from './BookCard';

const NewBook = () => {
  const [newBook, setNewBook] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchNewBook = async () => {
      try {
        setLoading(true);

        // ✅ ใช้ .env ถ้ามี หรือ fallback เป็น localhost:8080
        const apiUrl = process.env.REACT_APP_API_URL || "http://localhost:8080";

        // ✅ ดึงข้อมูลจาก backend ที่ route /api/v1/books/new
        const response = await fetch(`${apiUrl}/api/v1/books/new`);

        if (!response.ok) {
          throw new Error('Failed to fetch new books');
        }

        const data = await response.json();
        setNewBook(data);
        setError(null);

      } catch (err) {
        setError(err.message);
        console.error('Error fetching new books:', err);

      } finally {
        setLoading(false);
      }
    };

    // ✅ เรียก API ทันทีตอนหน้าโหลด
    fetchNewBook();
  }, []);

  // ⏳ Loading state
  if (loading) {
    return (
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
        <div className="text-center py-8 col-span-full text-gray-600">
          กำลังโหลดหนังสือใหม่...
        </div>
      </div>
    );
  }

  // ❌ Error state
  if (error) {
    return (
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
        <div className="text-center py-8 col-span-full text-red-600">
          Error: {error}
        </div>
      </div>
    );
  }

  // ✅ แสดงผลหนังสือใหม่
  return (
    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
      {newBook.map((book) => (
        <BookCard key={book.id} book={book} />
      ))}
    </div>
  );
};

export default NewBook;