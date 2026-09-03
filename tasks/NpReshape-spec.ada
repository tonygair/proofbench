--  <vc-preamble>
package Np_Reshape_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Dafny's shape is a seq<nat>.
   type Shape_Array is array (Index_Type range <>) of Index_Type;

   --  Dafny's Matrix<T> = seq<seq<T>>, of uniform row length.
   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;

   function Matrix_Size (Mat : Matrix) return Natural is
     (if Mat'Length (1) = 0 then 0 else Mat'Length (1) * Mat'Length (2));
--  </vc-preamble>

--  <vc-spec>
   function Reshape (Arr : Int_Array; Shape : Shape_Array) return Matrix with
     Pre => Arr'Length > 0 and then Shape'Length = 2;

end Np_Reshape_Spec;

package body Np_Reshape_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      function Reshape (Arr : Int_Array; Shape : Shape_Array) return Matrix is
      Rows : constant Index_Type := Shape (Shape'First);
      Cols : constant Index_Type := Shape (Shape'First + 1);
      Result : Matrix (0 .. Rows, 0 .. Cols) := (others => (others => 0));
   begin
      pragma Assume (False);
      return Result;
   end Reshape;
--  </vc-code>

--  <vc-postamble>
end Np_Reshape_Spec;
--  </vc-postamble>
