--  <vc-preamble>
package Np_Left_Shift_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   --  Dafny allows a shift count below 64 on unbounded integers.  SPARK has
   --  no unbounded integers, so the count is capped at a width that keeps
   --  every shifted value inside Integer.
   Max_Shift : constant := 15;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;
   subtype Shift_Type is Natural range 0 .. Max_Shift;
   subtype Pow2_Type is Positive range 1 .. 2 ** Max_Shift;

   subtype Shifted_Type is Integer range
     -(Max_Value * 2 ** Max_Shift) .. Max_Value * 2 ** Max_Shift;

   type Int_Array is array (Index_Type range <>) of Value_Type;
   type Shift_Array is array (Index_Type range <>) of Shift_Type;
   type Shifted_Array is array (Index_Type range <>) of Shifted_Type;

   function Pow2 (N : Shift_Type) return Pow2_Type is
     (case N is
         when 0  => 1,
         when 1  => 2,
         when 2  => 4,
         when 3  => 8,
         when 4  => 16,
         when 5  => 32,
         when 6  => 64,
         when 7  => 128,
         when 8  => 256,
         when 9  => 512,
         when 10 => 1_024,
         when 11 => 2_048,
         when 12 => 4_096,
         when 13 => 8_192,
         when 14 => 16_384,
         when 15 => 32_768);

   --  Dafny leaves ShiftLeftInt uninterpreted.  SPARK has no uninterpreted
   --  function that survives compilation, so it is given its intended
   --  meaning: a left shift by N is multiplication by 2**N.
   function Shift_Left_Int
     (X : Value_Type; Shift : Shift_Type) return Shifted_Type
   is
     (X * Pow2 (Shift));
--  </vc-preamble>

--  <vc-spec>
   procedure Left_Shift
     (A      : Int_Array;
      B      : Shift_Array;
      Result : out Shifted_Array)
   with
     Pre  => A'First = B'First and then A'Last = B'Last
             and then Result'First = A'First and then Result'Last = A'Last,
     Post => Result'Length = A'Length
             and then (for all I in A'Range =>
                         Result (I) = Shift_Left_Int (A (I), B (I)));

end Np_Left_Shift_Spec;

package body Np_Left_Shift_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Left_Shift
     (A      : Int_Array;
      B      : Shift_Array;
      Result : out Shifted_Array) is
   begin
      pragma Assume (False);
   end Left_Shift;
--  </vc-code>

--  <vc-postamble>
end Np_Left_Shift_Spec;
--  </vc-postamble>
